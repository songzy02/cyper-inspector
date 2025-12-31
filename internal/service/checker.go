// file: internal/service/checker.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"cyber-inspector/internal/agent"
	"cyber-inspector/internal/config"
	"cyber-inspector/internal/mailer"
	"cyber-inspector/internal/model"
	"cyber-inspector/internal/repository"
)

// Checker 巡检服务
type Checker struct {
	repo   *repository.Repository
	client *agent.Client
	cancel context.CancelFunc
	mu     sync.Mutex
	wg     sync.WaitGroup
}

// NewChecker 创建巡检服务
func NewChecker(repo *repository.Repository, client *agent.Client) *Checker {
	return &Checker{
		repo:   repo,
		client: client,
	}
}

// Start 启动巡检服务
func (c *Checker) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	log.Println("【巡检服务】已启动")

	// 立即执行一次巡检
	go c.batchCheck()

	// 启动定时巡检
	ticker := time.NewTicker(config.Conf.Check.Interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				go c.batchCheck()
			}
		}
	}()
}

// Stop 停止巡检服务
func (c *Checker) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	c.wg.Wait()
	log.Println("【巡检服务】已停止")
}

// batchCheck 批量巡检
func (c *Checker) batchCheck() {
	start := time.Now()
	log.Println("【批量巡检】开始...")

	agents, err := c.repo.GetActiveAgents()
	if err != nil {
		log.Printf("【批量巡检】获取节点列表失败: %v", err)
		return
	}

	if len(agents) == 0 {
		log.Println("【批量巡检】没有活跃的节点需要巡检")
		return
	}

	// 使用信号量控制并发数
	semaphore := make(chan struct{}, config.Conf.Check.MaxConcurrent)
	results := make(chan *InspectionResult, len(agents))

	// 启动巡检任务
	for _, agent := range agents {
		c.wg.Add(1)
		go c.checkAgent(agent, semaphore, results)
	}

	// 等待所有任务完成
	go func() {
		c.wg.Wait()
		close(results)
	}()

	// 处理结果
	c.processResults(results, start)
}

// InspectionResult 巡检结果
type InspectionResult struct {
	Agent      *model.Agent
	Inspection *model.Inspection
	Error      error
	Duration   time.Duration
}

// checkAgent 巡检单个节点
func (c *Checker) checkAgent(agent model.Agent, semaphore chan struct{}, results chan<- *InspectionResult) {
	defer c.wg.Done()

	// 获取信号量
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	start := time.Now()
	result := &InspectionResult{
		Agent:    &agent,
		Duration: 0,
	}

	// 重试机制
	for i := 0; i < config.Conf.Check.RetryTimes; i++ {
		inspection, err := c.client.Pull(agent)
		if err == nil {
			result.Inspection = inspection
			result.Duration = time.Since(start)
			break
		}

		result.Error = err
		if i < config.Conf.Check.RetryTimes-1 {
			time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
		}
	}

	results <- result
}

// processResults 处理巡检结果
func (c *Checker) processResults(results <-chan *InspectionResult, startTime time.Time) {
	var successCount, failedCount int

	for result := range results {
		if result.Error != nil {
			failedCount++
			log.Printf("【巡检失败】节点: %s, 错误: %v", result.Agent.Name, result.Error)
			continue
		}

		successCount++

		// 保存巡检结果
		if err := c.repo.SaveInspection(result.Inspection); err != nil {
			log.Printf("【保存失败】节点: %s, 错误: %v", result.Agent.Name, err)
			continue
		}

		// 处理告警
		c.processAlert(result.Inspection, result.Agent)

		// 更新Agent最后巡检时间
		c.repo.UpdateAgentStatus(result.Agent.ID, model.AgentOnline)

		log.Printf("【巡检成功】节点: %s, 级别: %s, 耗时: %v",
			result.Agent.Name, result.Inspection.Level, result.Duration)
	}

	elapsed := time.Since(startTime)
	log.Printf("【批量巡检】完成 %d 个节点, 成功: %d, 失败: %d, 耗时: %v",
		successCount+failedCount, successCount, failedCount, elapsed)
}

// processAlert 处理告警
func (c *Checker) processAlert(inspection *model.Inspection, agent *model.Agent) {
	log.Printf("[AlertDebug] 进入processAlert: agent=%s, level=%s, alertEnabled=%v",
		agent.Name, inspection.Level, config.Conf.Alert.Enabled)

	if !config.Conf.Alert.Enabled {
		log.Printf("[AlertDebug] 告警总开关关闭，直接返回")
		return
	}

	if inspection.Level != model.LevelCritical {
		log.Printf("[AlertDebug] 非CRITICAL，不告警")
		return
	}

	// 告警冷却检查
	if !c.canSendAlert(agent.ID, inspection.Level) {
		log.Printf("[AlertDebug] 冷却中，跳过")
		return
	}

	// 创建告警记录
	alert := &model.Alert{
		AgentID:      agent.ID,
		InspectionID: inspection.ID,
		Level:        inspection.Level,
		Title:        fmt.Sprintf("%s - %s", agent.Name, inspection.Level),
		Summary:      fmt.Sprintf("节点 %s 出现 %s 级别告警", agent.Name, inspection.Level),
	}

	// 解析分析结果
	var analysis struct {
		Summary string   `json:"summary"`
		Details []string `json:"details"`
		Plan    string   `json:"plan"`
	}
	if err := json.Unmarshal([]byte(inspection.Analysis), &analysis); err == nil {
		alert.Summary = analysis.Summary
		alert.Solution = analysis.Plan
		if len(analysis.Details) > 0 {
			alert.Details = analysis.Details[0]
		}
	}

	if err := c.repo.CreateAlert(alert); err != nil {
		log.Printf("【告警创建失败】节点: %s, 错误: %v", agent.Name, err)
		return
	}

	log.Printf("[AlertDebug] 告警已入库，ID=%d", alert.ID)

	// 发送邮件告警
	if config.Conf.Mail.Enabled && config.Conf.Mail.Host != "" {
		c.sendAlertMail(alert, agent)
	}
}

// 告警冷却缓存
var alertCooldown = make(map[string]time.Time)

// canSendAlert 检查是否可以发送告警
func (c *Checker) canSendAlert(agentID uint64, level model.InspectionLevel) bool {
	key := fmt.Sprintf("%d-%s", agentID, level)

	c.mu.Lock()
	defer c.mu.Unlock()

	if lastAlert, ok := alertCooldown[key]; ok {
		if time.Since(lastAlert) < config.Conf.Alert.Cooldown {
			return false
		}
	}

	alertCooldown[key] = time.Now()
	return true
}

// sendAlertMail 发送邮件告警
func (c *Checker) sendAlertMail(alert *model.Alert, agent *model.Agent) {
	sender := mailer.NewSender(
		config.Conf.Mail.Host,
		config.Conf.Mail.Port,
		config.Conf.Mail.User,
		config.Conf.Mail.Pass,
		config.Conf.Mail.To,
	)

	subject := fmt.Sprintf("%s %s - %s",
		config.Conf.Mail.SubjectPrefix,
		alert.Level,
		agent.Name,
	)

	body := fmt.Sprintf(`【Cyber Inspector 告警】

节点：%s
IP地址：%s
告警级别：%s
告警时间：%s
告警摘要：%s
解决方案：%s

请及时处理！
`,
		agent.Name,
		agent.IP,
		alert.Level,
		alert.CreatedAt.Format("2006-01-02 15:04:05"),
		alert.Summary,
		alert.Solution,
	)

	if err := sender.Send(subject, body); err != nil {
		log.Printf("【邮件发送失败】节点: %s, 错误: %v", agent.Name, err)
	} else {
		log.Printf("【邮件已发送】节点: %s", agent.Name)
		// 更新通知状态
		alert.Notified = true
		if err := c.repo.UpdateAlertStatus(alert.ID, model.AlertStatus(model.AlertPending)); err != nil {
			log.Printf("【更新通知状态失败】节点: %s, 错误: %v", agent.Name, err)
		}
	}
}

// BatchCheck 手动触发巡检
func (c *Checker) BatchCheck() {
	go c.batchCheck()
	log.Println("【手动巡检】已触发")
}
