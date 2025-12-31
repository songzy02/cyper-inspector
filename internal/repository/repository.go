package repository

import (
	"cyber-inspector/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
)

// Repository 数据仓库
type Repository struct {
	db *gorm.DB
}

// New 创建仓库实例
func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser 创建用户
func (r *Repository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// GetUserByUsername 根据用户名获取用户
func (r *Repository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func (r *Repository) GetUserByID(id uint64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers 获取所有用户
func (r *Repository) ListUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Order("id desc").Find(&users).Error
	return users, err
}

// UpdateUser 更新用户
func (r *Repository) UpdateUser(user *model.User) error {
	return r.db.Model(&model.User{}).Where("id = ?", user.ID).Updates(user).Error
}

// UpdateUserPassword 更新用户密码
func (r *Repository) UpdateUserPassword(id uint64, password string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("password", password).Error
}

// UpdateUserStatus 更新用户状态
func (r *Repository) UpdateUserStatus(id uint64, enabled bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// UpdateUserLastLogin 更新最后登录时间
func (r *Repository) UpdateUserLastLogin(id uint64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("last_login", "NOW()").Error
}

// DeleteUser 删除用户
func (r *Repository) DeleteUser(id uint64) error {
	return r.db.Delete(&model.User{}, id).Error
}

// CreateLoginLog 创建登录日志
func (r *Repository) CreateLoginLog(log *model.LoginLog) error {
	return r.db.Create(log).Error
}

// CreateAgent 创建Agent
//
//	func (r *Repository) CreateAgent(agent *model.Agent) error {
//		return r.db.Create(agent).Error
//	}
//
// CreateAgent 创建Agent
func (r *Repository) CreateAgent(agent *model.Agent) error {
	// 把零时间换成 nil，避免 '0000-00-00'
	if agent.LastCheckAt != nil && agent.LastCheckAt.IsZero() {
		agent.LastCheckAt = nil
	}
	return r.db.Create(agent).Error
}

// ListAgents 获取所有Agent
func (r *Repository) ListAgents() ([]model.Agent, error) {
	var agents []model.Agent
	err := r.db.Order("id desc").Find(&agents).Error
	return agents, err
}

// GetActiveAgents 获取活跃的Agent
func (r *Repository) GetActiveAgents() ([]model.Agent, error) {
	var agents []model.Agent
	err := r.db.Where("enabled = ?", true).Find(&agents).Error
	return agents, err
}

// GetAgentByID 根据ID获取Agent
func (r *Repository) GetAgentByID(id uint64) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.First(&agent, id).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// UpdateAgent 更新Agent
func (r *Repository) UpdateAgent(agent *model.Agent) error {
	return r.db.Model(&model.Agent{}).Where("id = ?", agent.ID).Updates(agent).Error
}

// DeleteAgent 删除Agent
func (r *Repository) DeleteAgent(id uint64) error {
	return r.db.Delete(&model.Agent{}, id).Error
}

// UpdateAgentStatus 更新Agent状态
func (r *Repository) UpdateAgentStatus(id uint64, status model.AgentStatus) error {
	return r.db.Model(&model.Agent{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateCheckInterval 更新巡检间隔
func (r *Repository) UpdateCheckInterval(id uint64, seconds int) error {
	return r.db.Model(&model.Agent{}).Where("id = ?", id).Update("check_interval", seconds).Error
}

// SaveInspection 保存巡检记录
func (r *Repository) SaveInspection(inspection *model.Inspection) error {
	return r.db.Create(inspection).Error
}

// LatestInspections 获取最新巡检记录
func (r *Repository) LatestInspections() ([]model.Inspection, error) {
	var inspections []model.Inspection
	err := r.db.Raw(`
		SELECT i.* FROM inspections i
		INNER JOIN (
			SELECT agent_id, MAX(created_at) AS max_time 
			FROM inspections 
			GROUP BY agent_id
		) t ON i.agent_id = t.agent_id AND i.created_at = t.max_time
	`).Scan(&inspections).Error
	return inspections, err
}

// GetInspectionsByAgentID 获取指定Agent的巡检记录
func (r *Repository) GetInspectionsByAgentID(agentID uint64, limit int) ([]model.Inspection, error) {
	var inspections []model.Inspection
	err := r.db.Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Limit(limit).
		Find(&inspections).Error
	return inspections, err
}

// CreateAlert 创建告警记录
func (r *Repository) CreateAlert(alert *model.Alert) error {
	return r.db.Create(alert).Error
}

// GetAlerts 获取告警记录
func (r *Repository) GetAlerts(limit, offset int) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.Preload("Agent").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&alerts).Error
	return alerts, err
}

// GetAlertStats 获取告警统计
func (r *Repository) GetAlertStats(days int) (map[string]interface{}, error) {
	var stats struct {
		Total    int64 `json:"total"`
		Pending  int64 `json:"pending"`
		Critical int64 `json:"critical"`
		Warning  int64 `json:"warning"`
	}

	// 统计总数
	if err := r.db.Model(&model.Alert{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 统计待处理
	if err := r.db.Model(&model.Alert{}).Where("status = ?", model.AlertPending).Count(&stats.Pending).Error; err != nil {
		return nil, err
	}

	// 统计严重告警
	if err := r.db.Model(&model.Alert{}).Where("level = ?", model.LevelCritical).Count(&stats.Critical).Error; err != nil {
		return nil, err
	}

	// 统计警告
	if err := r.db.Model(&model.Alert{}).Where("level = ?", model.LevelWarning).Count(&stats.Warning).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":    stats.Total,
		"pending":  stats.Pending,
		"critical": stats.Critical,
		"warning":  stats.Warning,
	}, nil
}

// UpdateAlertStatus 更新告警状态
func (r *Repository) UpdateAlertStatus(id uint64, status model.AlertStatus) error {
	return r.db.Model(&model.Alert{}).Where("id = ?", id).Update("status", status).Error
}

// InitAdminUser 初始化管理员用户
func (r *Repository) InitAdminUser(username, password string) error {
	// 检查是否已存在管理员
	var count int64
	if err := r.db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // 已有管理员，不创建
	}

	// 创建默认管理员
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     model.RoleAdmin,
		Enabled:  true,
	}

	return r.CreateUser(admin)
}

// CheckPassword 验证密码
// CheckPassword 兼容 bcrypt 哈希 & 明文（仅调试阶段允许明文）
func CheckPassword(plainPwd, hash string) bool {
	// 如果长度不是 60 且不以 $2 开头，就当明文直接比对
	if len(hash) != 60 || !strings.HasPrefix(hash, "$2") {
		return plainPwd == hash
	}
	// 否则走 bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPwd))
	return err == nil
}
