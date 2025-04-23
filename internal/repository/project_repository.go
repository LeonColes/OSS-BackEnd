package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
)

// ProjectRepository 项目仓库接口
type ProjectRepository interface {
	// 项目基本操作
	CreateProject(ctx context.Context, project *entity.Project) error
	UpdateProject(ctx context.Context, project *entity.Project) error
	GetProjectByID(ctx context.Context, id uint64) (*entity.Project, error)
	ListProjects(ctx context.Context, query *dto.ProjectQuery) ([]*entity.Project, int64, error)
	GetUserProjects(ctx context.Context, userID uint64, query *dto.ProjectQuery) ([]*entity.Project, int64, error)
	DeleteProject(ctx context.Context, id uint64) error

	// 项目权限操作
	SetPermission(ctx context.Context, permission *entity.Permission) error
	RemovePermission(ctx context.Context, projectID, userID uint64) error
	GetProjectPermission(ctx context.Context, projectID, userID uint64) (*entity.Permission, error)
	ListProjectUsers(ctx context.Context, projectID uint64) ([]*entity.Permission, error)

	// 检查用户是否拥有项目权限
	CheckUserProjectRole(ctx context.Context, userID, projectID uint64, roles []string) (bool, error)
}

// projectRepository 项目仓库实现
type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓库
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// CreateProject 创建项目
func (r *projectRepository) CreateProject(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// UpdateProject 更新项目
func (r *projectRepository) UpdateProject(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Model(&entity.Project{}).Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"name":        project.Name,
			"description": project.Description,
			"status":      project.Status,
			"updated_at":  time.Now(),
		}).Error
}

// GetProjectByID 通过ID获取项目
func (r *projectRepository) GetProjectByID(ctx context.Context, id uint64) (*entity.Project, error) {
	var project entity.Project
	err := r.db.WithContext(ctx).
		Preload("Group").
		Preload("Creator").
		First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects 列出项目列表
func (r *projectRepository) ListProjects(ctx context.Context, query *dto.ProjectQuery) ([]*entity.Project, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Project{})

	// 构建查询条件
	if query.GroupID > 0 {
		db = db.Where("group_id = ?", query.GroupID)
	}

	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}

	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?",
			fmt.Sprintf("%%%s%%", query.Keyword),
			fmt.Sprintf("%%%s%%", query.Keyword))
	}

	// 获取总数
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	page := query.Page
	if page <= 0 {
		page = 1
	}

	size := query.Size
	if size <= 0 {
		size = 10
	}

	var projects []*entity.Project
	err = db.Preload("Group").
		Preload("Creator").
		Offset((page - 1) * size).
		Limit(size).
		Order("created_at DESC").
		Find(&projects).Error

	return projects, total, err
}

// GetUserProjects 获取用户参与的项目
func (r *projectRepository) GetUserProjects(ctx context.Context, userID uint64, query *dto.ProjectQuery) ([]*entity.Project, int64, error) {
	db := r.db.WithContext(ctx).Table("projects p").
		Joins("JOIN permissions pm ON p.id = pm.project_id").
		Joins("JOIN users u ON pm.user_id = u.id").
		Where("pm.user_id = ?", userID)

	// 构建查询条件
	if query.GroupID > 0 {
		db = db.Where("p.group_id = ?", query.GroupID)
	}

	if query.Status > 0 {
		db = db.Where("p.status = ?", query.Status)
	}

	if query.Keyword != "" {
		db = db.Where("p.name LIKE ? OR p.description LIKE ?",
			fmt.Sprintf("%%%s%%", query.Keyword),
			fmt.Sprintf("%%%s%%", query.Keyword))
	}

	// 获取总数
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	page := query.Page
	if page <= 0 {
		page = 1
	}

	size := query.Size
	if size <= 0 {
		size = 10
	}

	var projectIDs []uint64
	err = db.Select("p.id").
		Offset((page-1)*size).
		Limit(size).
		Order("p.created_at DESC").
		Pluck("p.id", &projectIDs).Error

	if err != nil {
		return nil, 0, err
	}

	if len(projectIDs) == 0 {
		return []*entity.Project{}, 0, nil
	}

	var projects []*entity.Project
	err = r.db.WithContext(ctx).
		Preload("Group").
		Preload("Creator").
		Where("id IN ?", projectIDs).
		Order("created_at DESC").
		Find(&projects).Error

	return projects, total, err
}

// DeleteProject 删除项目(逻辑删除)
func (r *projectRepository) DeleteProject(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Project{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     3, // 设置为删除状态
			"updated_at": time.Now(),
		}).Error
}

// SetPermission 设置项目权限
func (r *projectRepository) SetPermission(ctx context.Context, permission *entity.Permission) error {
	// 先尝试更新，如果不存在则创建
	result := r.db.WithContext(ctx).Model(&entity.Permission{}).
		Where("project_id = ? AND user_id = ?", permission.ProjectID, permission.UserID).
		Updates(map[string]interface{}{
			"role":       permission.Role,
			"expire_at":  permission.ExpireAt,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// 不存在，则创建
		return r.db.WithContext(ctx).Create(permission).Error
	}

	return nil
}

// RemovePermission 移除项目权限
func (r *projectRepository) RemovePermission(ctx context.Context, projectID, userID uint64) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&entity.Permission{}).Error
}

// GetProjectPermission 获取用户的项目权限
func (r *projectRepository) GetProjectPermission(ctx context.Context, projectID, userID uint64) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&permission).Error

	if err != nil {
		return nil, err
	}

	return &permission, nil
}

// ListProjectUsers 列出项目用户及其权限
func (r *projectRepository) ListProjectUsers(ctx context.Context, projectID uint64) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Granter").
		Where("project_id = ?", projectID).
		Find(&permissions).Error

	return permissions, err
}

// CheckUserProjectRole 检查用户是否拥有项目特定角色
func (r *projectRepository) CheckUserProjectRole(ctx context.Context, userID, projectID uint64, roles []string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.Permission{}).
		Where("project_id = ? AND user_id = ?", projectID, userID)

	if len(roles) > 0 {
		query = query.Where("role IN ?", roles)
	}

	err := query.Count(&count).Error
	return count > 0, err
}
