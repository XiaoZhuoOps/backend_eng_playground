/* 
# Context
你在互联网公司负责开发一个 ToC 的基于 Golang Ginex + Gorm 的 Web服务
该服务的 QPS 峰值达百万级 其核心逻辑链路为 Req -> LocalCache -> Redis -> Mysql
Mysql 中存储记录的表结构为
type AnalyticsWeb struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	WebCode        string         `gorm:"column:web_code;not null" json:"web_code"`
	Name           string         `gorm:"column:name;not null" json:"name"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP(1)" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	Version        int32          `gorm:"column:version;not null" json:"version"`
	Extra          string         `gorm:"column:extra;not null;default:{}" json:"extra"`
	EventSetupMode int32          `gorm:"column:event_setup_mode;not null" json:"event_setup_mode"`
}
其中 Extra 是一个 JSON 结构体 里面包含一些用户的配置
该服务只提供读接口 且只在机房B serving
另外一个ToB服务 QPS个位数 用于写/更新 mysql 且只在机房 A serving
目前的现状是 某个用户修改 mysql后 通过 binlog将修改从机房 A同步到机房B 但没有主动地清理/更新 redis 和 local cache
导致ToC 接口存在 30min 的数据不一致延迟（Redis 过期时间）

# Task
你需要解决这个问题，将 ToC 服务的 数据不一致延迟降低到合适的时间比如 5min 内
*/