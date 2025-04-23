package entity

// CasbinRule Casbin规则实体
type CasbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:100;uniqueIndex:unique_policy;comment:策略类型"`
	V0    string `gorm:"size:100;uniqueIndex:unique_policy;comment:角色/用户"`
	V1    string `gorm:"size:100;uniqueIndex:unique_policy;comment:域/租户"`
	V2    string `gorm:"size:100;uniqueIndex:unique_policy;comment:资源"`
	V3    string `gorm:"size:100;uniqueIndex:unique_policy;comment:操作"`
	V4    string `gorm:"size:100;uniqueIndex:unique_policy;comment:预留字段"`
	V5    string `gorm:"size:100;uniqueIndex:unique_policy;comment:预留字段"`
}

// TableName 表名
func (CasbinRule) TableName() string {
	return "casbin_rule"
}
