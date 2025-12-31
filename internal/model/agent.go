package model

import "time"

// AgentStatus Agent状态
type AgentStatus string

const (
	AgentOnline  AgentStatus = "online"
	AgentOffline AgentStatus = "offline"
	AgentUnknown AgentStatus = "unknown"
)

// Agent Agent节点模型
type Agent struct {
	ID            uint64      `gorm:"primaryKey" json:"id"`
	Name          string      `gorm:"size:64;not null" json:"name"`          // 节点名称
	IP            string      `gorm:"size:15;not null" json:"ip"`            // IP地址
	URL           string      `gorm:"size:128;unique;not null" json:"url"`   // Agent URL
	Enabled       bool        `gorm:"default:true" json:"enabled"`           // 是否启用
	CheckInterval int         `gorm:"default:300" json:"check_interval"`     // 巡检间隔（秒）
	APIKey        string      `gorm:"size:255" json:"-"`                     // API密钥
	Status        AgentStatus `gorm:"size:20;default:unknown" json:"status"` // 节点状态
	//LastCheckAt   time.Time   `json:"last_check_at"`
	LastCheckAt *time.Time `gorm:"default:null;column:last_check_at" json:"last_check_at,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (Agent) TableName() string {
	return "agents"
}

// InspectionLevel 巡检级别
type InspectionLevel string

const (
	LevelOK       InspectionLevel = "OK"
	LevelWarning  InspectionLevel = "WARNING"
	LevelCritical InspectionLevel = "CRITICAL"
)

// Inspection 巡检记录模型
type Inspection struct {
	ID             uint64          `gorm:"primaryKey" json:"id"`
	AgentID        uint64          `gorm:"not null;index" json:"agent_id"`       // Agent ID
	Hostname       string          `gorm:"size:64;not null" json:"hostname"`     // 主机名
	IP             string          `gorm:"size:15;not null" json:"ip"`           // IP地址
	RawData        string          `gorm:"type:text" json:"raw_data"`            // 原始数据
	Analysis       string          `gorm:"type:text" json:"analysis"`            // 分析结果
	Alert          bool            `gorm:"default:false" json:"alert"`           // 是否告警
	Level          InspectionLevel `gorm:"size:16;default:OK" json:"level"`      // 告警级别
	CPUUsed        float64         `gorm:"type:decimal(5,2)" json:"cpu_used"`    // CPU使用率
	MemoryUsed     float64         `gorm:"type:decimal(5,2)" json:"memory_used"` // 内存使用率
	DiskUsed       float64         `gorm:"type:decimal(5,2)" json:"disk_used"`   // 磁盘使用率
	LoadAvg        float64         `gorm:"type:decimal(5,2)" json:"load_avg"`    // 平均负载
	PingLoss       float64         `gorm:"type:decimal(5,2)" json:"ping_loss"`   // 网络丢包率
	JournalErr1h   int             `json:"journal_err_1h"`                       // 1小时内错误日志数
	ProcessCount   int             `json:"process_count"`                        // 进程数
	TCPConnections int             `json:"tcp_connections"`                      // TCP连接数
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	Agent          Agent           `gorm:"foreignKey:AgentID" json:"-"`
}

// TableName 表名
func (Inspection) TableName() string {
	return "inspections"
}

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertPending    AlertStatus = "pending"
	AlertProcessing AlertStatus = "processing"
	AlertResolved   AlertStatus = "resolved"
	AlertIgnored    AlertStatus = "ignored"
)

// Alert 告警记录模型
type Alert struct {
	ID           uint64          `gorm:"primaryKey" json:"id"`
	AgentID      uint64          `gorm:"not null;index" json:"agent_id"`        // Agent ID
	InspectionID uint64          `gorm:"not null;index" json:"inspection_id"`   // 巡检记录ID
	Level        InspectionLevel `gorm:"size:16;not null" json:"level"`         // 告警级别
	Title        string          `gorm:"size:255;not null" json:"title"`        // 告警标题
	Summary      string          `gorm:"type:text" json:"summary"`              // 告警摘要
	Details      string          `gorm:"type:text" json:"details"`              // 详细信息
	Solution     string          `gorm:"type:text" json:"solution"`             // 解决方案
	Status       AlertStatus     `gorm:"size:20;default:pending" json:"status"` // 告警状态
	Notified     bool            `gorm:"default:false" json:"notified"`         // 是否已通知
	//ResolvedAt   time.Time       `json:"resolved_at"`                           // 解决时间
	ResolvedAt *time.Time `gorm:"default:null;column:resolved_at" json:"resolved_at,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Agent      Agent      `gorm:"foreignKey:AgentID" json:"-"`
	Inspection Inspection `gorm:"foreignKey:InspectionID" json:"-"`
}

// TableName 表名
func (Alert) TableName() string {
	return "alerts"
}
