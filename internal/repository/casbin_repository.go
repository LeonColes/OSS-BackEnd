package repository

import (
	"gorm.io/gorm"
)

// CasbinRepository Casbin规则仓库接口
type CasbinRepository interface {
	// DeleteRoleRules 删除与角色相关的所有规则
	DeleteRoleRules(tx *gorm.DB, roleCode string) error
}

// casbinRepository Casbin规则仓库实现
type casbinRepository struct {
	db *gorm.DB
}

// NewCasbinRepository 创建Casbin规则仓库
func NewCasbinRepository(db *gorm.DB) CasbinRepository {
	return &casbinRepository{
		db: db,
	}
}

// DeleteRoleRules 删除与角色相关的所有规则
func (r *casbinRepository) DeleteRoleRules(tx *gorm.DB, roleCode string) error {
	// 删除角色作为主体的规则 (p_type = 'p')
	err := tx.Table("casbin_rule").
		Where("p_type = ? AND v0 = ?", "p", roleCode).
		Delete(nil).Error
	if err != nil {
		return err
	}

	// 删除角色关联规则 (p_type = 'g')
	err = tx.Table("casbin_rule").
		Where("p_type = ? AND v1 = ?", "g", roleCode).
		Delete(nil).Error
	if err != nil {
		return err
	}

	return nil
}
