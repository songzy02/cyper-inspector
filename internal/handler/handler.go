package handler

import (
	"cyber-inspector/internal/auth"
	"cyber-inspector/internal/config"
	"cyber-inspector/internal/model"
	"cyber-inspector/internal/repository"
	"cyber-inspector/internal/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Login 用户登录
func Login(repo *repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查找用户
		user, err := repo.GetUserByUsername(req.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		if !user.Enabled {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "账户已被禁用"})
			return
		}
		// 验证密码前打印调试
		log.Printf("[LoginDebug] 用户输入 -> username=%s, password=%s", req.Username, req.Password)
		// 验证密码
		if !repository.CheckPassword(req.Password, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		// 生成 JWT Token
		//token, err := GenerateToken(user)
		token, err := auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}

		// 不返回密码
		user.Password = ""

		c.JSON(http.StatusOK, LoginResponse{
			Token: token,
			User:  user,
		})
	}
}

// Logout 用户登出
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
	}
}

// ListAgents 获取节点列表
func ListAgents(repo *repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		agents, err := repo.ListAgents()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 获取最新巡检记录
		inspections, _ := repo.LatestInspections()
		statusMap := make(map[uint64]model.Inspection)
		for _, ins := range inspections {
			statusMap[ins.AgentID] = ins
		}

		c.JSON(http.StatusOK, gin.H{
			"agents":    agents,
			"statusMap": statusMap,
		})
	}
}

// CreateAgentRequest 创建节点请求
type CreateAgentRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=64"`
	IP            string `json:"ip" binding:"required,ip"`
	URL           string `json:"url" binding:"required,url"`
	CheckInterval int    `json:"check_interval" binding:"min=30"`
	Enabled       bool   `json:"enabled"`
}

// CreateAgent 创建节点
func CreateAgent(repo *repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateAgentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		agent := &model.Agent{
			Name:          req.Name,
			IP:            req.IP,
			URL:           req.URL,
			CheckInterval: req.CheckInterval,
			Enabled:       req.Enabled,
		}

		if err := repo.CreateAgent(agent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, agent)
	}
}

// UpdateAgent 更新节点
func UpdateAgent(repo *repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

		var req CreateAgentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		agent, err := repo.GetAgentByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
			return
		}

		agent.Name = req.Name
		agent.IP = req.IP
		agent.URL = req.URL
		agent.CheckInterval = req.CheckInterval
		agent.Enabled = req.Enabled

		if err := repo.UpdateAgent(agent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, agent)
	}
}

// DeleteAgent 删除节点
func DeleteAgent(repo *repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

		if err := repo.DeleteAgent(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}

// UpdateIntervalRequest 更新间隔请求
type UpdateIntervalRequest struct {
	Seconds int `json:"seconds" binding:"min=30"`
}

// UpdateInterval 更新巡检间隔
func UpdateInterval(repo *repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

		var req UpdateIntervalRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "周期必须≥30秒"})
			return
		}

		if err := repo.UpdateCheckInterval(id, req.Seconds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id, "interval": req.Seconds})
	}
}

// TriggerCheck 触发巡检
func TriggerCheck(checker *service.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		checker.BatchCheck()
		c.JSON(http.StatusOK, gin.H{"message": "巡检已触发"})
	}
}

// GetStatus 获取巡检状态
type StatusResponse struct {
	IsRunning     bool      `json:"is_running"`
	LastCheckTime time.Time `json:"last_check_time"`
	NextCheckTime time.Time `json:"next_check_time"`
	TotalNodes    int       `json:"total_nodes"`
}

func GetStatus(checker *service.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, GetStatusResponse())
	}
}

// GetStatusResponse 获取状态响应
func GetStatusResponse() StatusResponse {
	return StatusResponse{
		IsRunning:     true,
		LastCheckTime: time.Now().Add(-config.Conf.Check.Interval),
		NextCheckTime: time.Now().Add(config.Conf.Check.Interval),
		TotalNodes:    0,
	}
}
