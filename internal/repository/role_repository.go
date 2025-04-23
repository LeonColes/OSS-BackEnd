package repository

import (
	"context"

	"gorm.io/gorm"

	"oss-backend/internal/model/entity"
)

// RoleRepository 角色仓库接口
type RoleRepository interface {
	// Create 创建角色
	Create(ctx context.Context, role *entity.Role) error
	// Update 更新角色
	Update(ctx context.Context, role *entity.Role) error
	// Delete 删除角色
	Delete(ctx context.Context, id uint) error
	// GetByID 根据ID获取角色
	GetByID(ctx context.Context, id uint) (*entity.Role, error)
	// GetByCode 根据Code获取角色
	GetByCode(ctx context.Context, code string) (*entity.Role, error)
	// List 获取角色列表
	List(ctx context.Context, name string, status int, page, size int) ([]*entity.Role, int64, error)
	// InitSystemRoles 初始化系统角色
	InitSystemRoles(ctx context.Context) error
}

// roleRepository 角色仓库实现
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{
		db: db,
	}
}

// Create 创建角色
func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Update 更新角色
func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Model(&entity.Role{}).Where("id = ?", role.ID).Updates(role).Error
}

// Delete 删除角色
func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

// GetByID 根据ID获取角色
func (r *roleRepository) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByCode 根据Code获取角色
func (r *roleRepository) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// List 获取角色列表
func (r *roleRepository) List(ctx context.Context, name string, status int, page, size int) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Role{})

	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}

	if status != -1 {
		db = db.Where("status = ?", status)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if page > 0 && size > 0 {
		offset := (page - 1) * size
		db = db.Offset(offset).Limit(size)
	}

	err = db.Order("id DESC").Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// InitSystemRoles 初始化系统角色
func (r *roleRepository) InitSystemRoles(ctx context.Context) error {
	// 预定义的系统角色
	systemRoles := []*entity.Role{
		{
			Name:        "群组管理员",
			Description: "群组管理员",
			Code:        entity.RoleGroupAdmin,
			Status:      1,
			IsSystem:    true,
		},
		{
			Name:        "项目管理员",
			Description: "项目管理员",
			Code:        entity.RoleProjectAdmin,
			Status:      1,
			IsSystem:    true,
		},
		{
			Name:        "普通成员",
			Description: "普通成员",
			Code:        entity.RoleMember,
			Status:      1,
			IsSystem:    true,
		},
		{
			Name:        "上传者",
			Description: "可以上传文件的用户",
			Code:        entity.RoleUploader,
			Status:      1,
			IsSystem:    true,
		},
		{
			Name:        "只读用户",
			Description: "只能查看文件的用户",
			Code:        entity.RoleReader,
			Status:      1,
			IsSystem:    true,
		},
	}

	// 事务处理
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, role := range systemRoles {
			// 检查是否已存在
			var count int64
			err := tx.Model(&entity.Role{}).Where("code = ?", role.Code).Count(&count).Error
			if err != nil {
				return err
			}

			// 不存在则创建
			if count == 0 {
				if err := tx.Create(role).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
