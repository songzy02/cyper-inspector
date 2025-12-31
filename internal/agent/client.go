package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cyber-inspector/internal/model"
)

// Client HTTP客户端
type Client struct {
	http *http.Client
}

// NewClient 创建客户端
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Pull 拉取Agent数据
func (c *Client) Pull(agent model.Agent) (*model.Inspection, error) {
	start := time.Now()

	url := agent.URL + "/inspect"
	resp, err := c.http.Get(url)
	if err != nil {
		log.Printf("【Agent 网络不通】url=%s elapsed=%v err=%v", url, time.Since(start), err)
		return &model.Inspection{
			AgentID:  agent.ID,
			Hostname: agent.Name,
			IP:       agent.IP,
			Alert:    true,
			Level:    model.LevelCritical,
			Analysis: `{"summary":"Agent unreachable"}`,
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("【Agent 原始回包】url=%s status=%d elapsed=%v len=%d",
		url, resp.StatusCode, time.Since(start), len(body))

	// 解析响应
	var response struct {
		Hostname string          `json:"hostname"`
		RawData  json.RawMessage `json:"raw_data"`
		Analysis json.RawMessage `json:"analysis"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return &model.Inspection{
			AgentID:  agent.ID,
			Hostname: agent.Name,
			IP:       agent.IP,
			Alert:    true,
			Level:    model.LevelCritical,
			Analysis: `{"summary":"bad response format"}`,
		}, nil
	}

	// 解析分析结果
	var content struct {
		Alert bool                  `json:"alert"`
		Level model.InspectionLevel `json:"level"`
	}
	_ = json.Unmarshal(response.Analysis, &content)

	// 创建巡检记录
	inspection := &model.Inspection{
		AgentID:  agent.ID,
		Hostname: response.Hostname,
		IP:       agent.IP,
		RawData:  string(response.RawData),
		Analysis: string(response.Analysis),
		Alert:    content.Alert,
		Level:    content.Level,
	}

	// 解析系统指标
	var sysInfo struct {
		CPUUsed   string `json:"cpu_used"`
		MemUsed   string `json:"mem_used"`
		DiskAlert string `json:"disk_alert"`
		LoadAvg   string `json:"cpu_load"`
		PingLoss  string `json:"ping_loss"`
	}
	if err := json.Unmarshal(response.RawData, &sysInfo); err == nil {
		// 解析数值（简化处理）
		fmt.Sscanf(sysInfo.CPUUsed, "%f", &inspection.CPUUsed)
		fmt.Sscanf(sysInfo.MemUsed, "%f", &inspection.MemoryUsed)
		fmt.Sscanf(sysInfo.LoadAvg, "%f", &inspection.LoadAvg)
		fmt.Sscanf(sysInfo.PingLoss, "%f", &inspection.PingLoss)
	}

	return inspection, nil
}
