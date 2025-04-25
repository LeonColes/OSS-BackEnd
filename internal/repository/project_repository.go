package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/utils"
)

// ProjectRepository 项目仓库接口
type ProjectRepository interface {
	// 事务支持
	WithTx(tx *gorm.DB) ProjectRepository

	// 基础CRUD
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id string) (*entity.Project, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id string) error

	// 查询方法
	List(ctx context.Context, req *dto.ProjectListRequest) ([]entity.Project, int64, error)
	GetByGroupID(ctx context.Context, groupID string) ([]entity.Project, error)
	GetUserProjects(ctx context.Context, userID string) ([]entity.Project, error)

	// 权限相关
	CreateProjectMember(ctx context.Context, member *entity.ProjectMember) error
	GetProjectMember(ctx context.Context, projectID, userID string) (*entity.ProjectMember, error)
	UpdateProjectMember(ctx context.Context, member *entity.ProjectMember) error
	RemoveProjectMember(ctx context.Context, projectID, userID string) error
	ListProjectMembers(ctx context.Context, projectID string, page, size int) ([]entity.ProjectMember, int64, error)
	CheckUserProjectRole(ctx context.Context, userID, projectID string, role string) (bool, error)
	CheckUserInProject(ctx context.Context, userID, projectID string) (bool, error)
	AddProjectPermission(ctx context.Context, permission *entity.Permission) error
	GetProjectPermission(ctx context.Context, projectID, userID string) (*entity.Permission, error)
	UpdateProjectPermission(ctx context.Context, permission *entity.Permission) error
	RemoveProjectPermission(ctx context.Context, projectID, userID string) error
	ListProjectPermissions(ctx context.Context, projectID string, page, size int) ([]entity.Permission, int64, error)
}

// projectRepository 项目仓库实现
type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓库
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

// Create 创建项目
func (r *projectRepository) Create(ctx context.Context, project *entity.Project) error {
	if project.ID == "" {
		project.ID = utils.GenerateProjectID()
	}
	return r.db.WithContext(ctx).Create(project).Error
}

// GetByID 根据ID获取项目
func (r *projectRepository) GetByID(ctx context.Context, id string) (*entity.Project, error) {
	var project entity.Project
	err := r.db.WithContext(ctx).Where("id = ?", id).Preload("Group").First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

// Update 更新项目
func (r *projectRepository) Update(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete 删除项目
func (r *projectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Project{}, "id = ?", id).Error
}

// List 获取项目列表
func (r *projectRepository) List(ctx context.Context, req *dto.ProjectListRequest) ([]entity.Project, int64, error) {
	var projects []entity.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Project{})

	// 条件筛选
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	if req.GroupID != "" {
		query = query.Where("group_id = ?", req.GroupID)
	}

	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if req.Page > 0 && req.PageSize > 0 {
		offset := (req.Page - 1) * req.PageSize
		query = query.Offset(offset).Limit(req.PageSize)
	}

	// 排序
	if req.SortBy != "" {
		order := req.SortBy
		if req.SortOrder == "desc" {
			order += " DESC"
		} else {
			order += " ASC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// 执行查询
	err = query.Preload("Group").Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// GetByGroupID 根据群组ID获取项目
func (r *projectRepository) GetByGroupID(ctx context.Context, groupID string) ([]entity.Project, error) {
	var projects []entity.Project
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&projects).Error
	return projects, err
}

// GetUserProjects 获取用户的项目
func (r *projectRepository) GetUserProjects(ctx context.Context, userID string) ([]entity.Project, error) {
	var projects []entity.Project
	err := r.db.WithContext(ctx).
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userID).
		Preload("Group").
		Find(&projects).Error
	return projects, err
}

// CreateProjectMember 添加项目成员
func (r *projectRepository) CreateProjectMember(ctx context.Context, member *entity.ProjectMember) error {
	if member.ID == "" {
		member.ID = utils.GenerateRecordID()
	}
	return r.db.WithContext(ctx).Create(member).Error
}

// GetProjectMember 获取项目成员
func (r *projectRepository) GetProjectMember(ctx context.Context, projectID, userID string) (*entity.ProjectMember, error) {
	var member entity.ProjectMember
	err := r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// UpdateProjectMember 更新项目成员
func (r *projectRepository) UpdateProjectMember(ctx context.Context, member *entity.ProjectMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// RemoveProjectMember 移除项目成员
func (r *projectRepository) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&entity.ProjectMember{}).Error
}

// ListProjectMembers 获取项目成员列表
func (r *projectRepository) ListProjectMembers(ctx context.Context, projectID string, page, size int) ([]entity.ProjectMember, int64, error) {
	var members []entity.ProjectMember
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ProjectMember{}).Where("project_id = ?", projectID)

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && size > 0 {
		offset := (page - 1) * size
		query = query.Offset(offset).Limit(size)
	}

	// 执行查询
	err = query.Preload("User").Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// CheckUserProjectRole 检查用户在项目中的角色
func (r *projectRepository) CheckUserProjectRole(ctx context.Context, userID, projectID string, role string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND role = ?", projectID, userID, role).
		Count(&count).Error

	return count > 0, err
}

// CheckUserInProject 检查用户是否在项目中
func (r *projectRepository) CheckUserInProject(ctx context.Context, userID, projectID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error

	return count > 0, err
}

// AddProjectPermission 添加项目权限
func (r *projectRepository) AddProjectPermission(ctx context.Context, permission *entity.Permission) error {
	if permission.ID == "" {
		permission.ID = utils.GenerateRecordID()
	}
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetProjectPermission 获取项目权限
func (r *projectRepository) GetProjectPermission(ctx context.Context, projectID, userID string) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// UpdateProjectPermission 更新项目权限
func (r *projectRepository) UpdateProjectPermission(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// RemoveProjectPermission 移除项目权限
func (r *projectRepository) RemoveProjectPermission(ctx context.Context, projectID, userID string) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&entity.Permission{}).Error
}

// ListProjectPermissions 获取项目权限列表
func (r *projectRepository) ListProjectPermissions(ctx context.Context, projectID string, page, size int) ([]entity.Permission, int64, error) {
	var permissions []entity.Permission
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Permission{}).Where("project_id = ?", projectID)

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && size > 0 {
		offset := (page - 1) * size
		query = query.Offset(offset).Limit(size)
	}

	// 执行查询
	err = query.Preload("User").Find(&permissions).Error
	if err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// WithTx 事务支持
func (r *projectRepository) WithTx(tx *gorm.DB) ProjectRepository {
	return &projectRepository{
		db: tx,
	}
}
