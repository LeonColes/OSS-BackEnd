package service

import (
	"errors"
	"strings"
	"time"

	"oss-backend/internal/model/entity"

	"gorm.io/gorm"
)

// 定义项目相关错误
var (
	ErrProjectNotFound     = errors.New("项目不存在")
	ErrProjectAlreadyExist = errors.New("项目已存在")
	ErrPathPrefixExists    = errors.New("存储路径前缀已存在")
	ErrNoProjectPermission = errors.New("无项目操作权限")
	ErrUserNotFound        = errors.New("用户不存在")
	ErrInvalidRole         = errors.New("无效的角色")
	ErrMemberAlreadyExists = errors.New("成员已存在")
	ErrMemberNotFound      = errors.New("成员不存在")
	ErrCannotRemoveSelf    = errors.New("不能移除自己")
	// ErrNoPermission 已在file_service.go中定义
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 创建项目
	CreateProject(groupID uint, name, description, pathPrefix string, creatorID uint) (*entity.Project, error)
	// 获取项目信息
	GetProject(id uint) (*entity.Project, error)
	// 更新项目信息
	UpdateProject(id uint, name, description string, userID uint) (*entity.Project, error)
	// 删除项目
	DeleteProject(id uint, userID uint) error
	// 获取群组下的所有项目
	GetGroupProjects(groupID uint, userID uint) ([]*entity.Project, error)
	// 获取用户有权限的所有项目
	GetUserProjects(userID uint) ([]*entity.Project, error)
	// 获取项目成员
	GetProjectMembers(projectID uint) ([]*entity.Permission, error)
	// 添加项目成员
	AddProjectMember(projectID, targetUserID uint, role string, userID uint) error
	// 移除项目成员
	RemoveProjectMember(projectID, targetUserID, userID uint) error
	// 更新项目成员角色
	UpdateProjectMemberRole(projectID, targetUserID uint, role string, userID uint) error
	// 检查用户是否有项目权限
	CheckProjectPermission(projectID, userID uint, requiredRole string) (bool, error)
}

type projectService struct {
	db *gorm.DB
}

// NewProjectService 创建项目服务实例
func NewProjectService(db *gorm.DB) ProjectService {
	return &projectService{db: db}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(groupID uint, name, description, pathPrefix string, creatorID uint) (*entity.Project, error) {
	// 检查用户是否有权限创建项目
	var count int64
	err := s.db.Model(&entity.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND role IN ?", groupID, creatorID, []string{entity.GroupRoleOwner, entity.GroupRoleAdmin}).
		Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, ErrNoProjectPermission
	}

	// 检查路径前缀是否已存在
	pathPrefix = strings.ToLower(strings.TrimSpace(pathPrefix))
	if pathPrefix == "" {
		// 默认使用项目名称的小写作为路径前缀
		pathPrefix = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	var existingProject entity.Project
	err = s.db.Where("group_id = ? AND path_prefix = ?", groupID, pathPrefix).First(&existingProject).Error
	if err == nil {
		return nil, ErrPathPrefixExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 创建项目
	project := &entity.Project{
		GroupID:     uint64(groupID),
		Name:        name,
		Description: description,
		PathPrefix:  pathPrefix,
		CreatorID:   uint64(creatorID),
		Status:      1, // 正常状态
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建项目
		if err := tx.Create(project).Error; err != nil {
			return err
		}

		// 添加创建者为项目管理员
		permission := &entity.Permission{
			UserID:    uint64(creatorID),
			ProjectID: project.ID,
			Role:      entity.ProjectRoleAdmin,
			GrantedBy: uint64(creatorID),
		}

		return tx.Create(permission).Error
	})

	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject 获取项目信息
func (s *projectService) GetProject(id uint) (*entity.Project, error) {
	var project entity.Project
	err := s.db.Preload("Group").Preload("Creator").First(&project, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return &project, nil
}

// UpdateProject 更新项目信息
func (s *projectService) UpdateProject(id uint, name, description string, userID uint) (*entity.Project, error) {
	// 检查项目是否存在
	var project entity.Project
	err := s.db.First(&project, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// 检查用户是否有权限更新项目
	hasPermission, err := s.CheckProjectPermission(id, userID, entity.ProjectRoleAdmin)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrNoProjectPermission
	}

	// 更新项目信息
	if name != "" {
		project.Name = name
	}
	if description != "" {
		project.Description = description
	}

	err = s.db.Save(&project).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(id uint, userID uint) error {
	// 检查项目是否存在
	var project entity.Project
	err := s.db.First(&project, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	// 检查用户是否有权限删除项目
	hasPermission, err := s.CheckProjectPermission(id, userID, entity.ProjectRoleAdmin)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrNoProjectPermission
	}

	// 软删除项目（使用GORM的软删除功能）
	return s.db.Delete(&project).Error
}

// GetGroupProjects 获取群组下的所有项目
func (s *projectService) GetGroupProjects(groupID uint, userID uint) ([]*entity.Project, error) {
	// 检查用户是否为群组成员
	var count int64
	err := s.db.Model(&entity.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, ErrNoPermission
	}

	// 获取群组下的所有项目
	var projects []*entity.Project
	err = s.db.Preload("Creator").
		Where("group_id = ? AND status = ?", groupID, 1).
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// GetUserProjects 获取用户有权限的所有项目
func (s *projectService) GetUserProjects(userID uint) ([]*entity.Project, error) {
	// 获取用户有权限的所有项目ID
	var permissions []*entity.Permission
	err := s.db.Where("user_id = ?", userID).Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	if len(permissions) == 0 {
		return []*entity.Project{}, nil
	}

	// 收集项目ID
	var projectIDs []uint64
	for _, perm := range permissions {
		projectIDs = append(projectIDs, perm.ProjectID)
	}

	// 获取项目详情
	var projects []*entity.Project
	err = s.db.Preload("Group").Preload("Creator").
		Where("id IN ? AND status = ?", projectIDs, 1).
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProjectMembers 获取项目成员
func (s *projectService) GetProjectMembers(projectID uint) ([]*entity.Permission, error) {
	// 获取项目所有权限
	var permissions []*entity.Permission
	err := s.db.Preload("User").Preload("Granter").
		Where("project_id = ?", projectID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddProjectMember 添加项目成员
func (s *projectService) AddProjectMember(projectID, targetUserID uint, role string, userID uint) error {
	// 检查项目是否存在
	var project entity.Project
	err := s.db.First(&project, projectID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	// 检查用户是否有权限添加成员
	hasPermission, err := s.CheckProjectPermission(projectID, userID, entity.ProjectRoleAdmin)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrNoProjectPermission
	}

	// 检查目标用户是否存在
	var targetUser entity.User
	err = s.db.First(&targetUser, targetUserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 检查角色是否有效
	validRoles := []string{entity.ProjectRoleAdmin, entity.ProjectRoleEditor, entity.ProjectRoleViewer}
	roleValid := false
	for _, validRole := range validRoles {
		if role == validRole {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return ErrInvalidRole
	}

	// 检查用户是否已经是项目成员
	var existingPerm entity.Permission
	err = s.db.Where("project_id = ? AND user_id = ?", projectID, targetUserID).First(&existingPerm).Error
	if err == nil {
		// 如果已存在，更新角色
		existingPerm.Role = role
		existingPerm.UpdatedAt = time.Now()
		err = s.db.Save(&existingPerm).Error
		if err != nil {
			return err
		}
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 添加新的权限记录
	permission := &entity.Permission{
		UserID:    uint64(targetUserID),
		ProjectID: uint64(projectID),
		Role:      role,
		GrantedBy: uint64(userID),
	}

	err = s.db.Create(permission).Error
	if err != nil {
		return err
	}

	return nil
}

// RemoveProjectMember 移除项目成员
func (s *projectService) RemoveProjectMember(projectID, targetUserID, userID uint) error {
	// 检查项目是否存在
	var project entity.Project
	err := s.db.First(&project, projectID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	// 检查用户是否有权限移除成员
	hasPermission, err := s.CheckProjectPermission(projectID, userID, entity.ProjectRoleAdmin)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrNoProjectPermission
	}

	// 不能移除自己
	if userID == targetUserID {
		return ErrCannotRemoveSelf
	}

	// 检查目标成员是否存在
	var permission entity.Permission
	err = s.db.Where("project_id = ? AND user_id = ?", projectID, targetUserID).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	// 删除权限记录
	return s.db.Delete(&permission).Error
}

// UpdateProjectMemberRole 更新项目成员角色
func (s *projectService) UpdateProjectMemberRole(projectID, targetUserID uint, role string, userID uint) error {
	// 检查项目是否存在
	var project entity.Project
	err := s.db.First(&project, projectID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	// 检查用户是否有权限更新成员角色
	hasPermission, err := s.CheckProjectPermission(projectID, userID, entity.ProjectRoleAdmin)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrNoProjectPermission
	}

	// 检查角色是否有效
	validRoles := []string{entity.ProjectRoleAdmin, entity.ProjectRoleEditor, entity.ProjectRoleViewer}
	roleValid := false
	for _, validRole := range validRoles {
		if role == validRole {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return ErrInvalidRole
	}

	// 检查目标成员是否存在
	var permission entity.Permission
	err = s.db.Where("project_id = ? AND user_id = ?", projectID, targetUserID).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	// 更新角色
	permission.Role = role
	permission.UpdatedAt = time.Now()
	permission.GrantedBy = uint64(userID)

	err = s.db.Save(&permission).Error
	if err != nil {
		return err
	}

	return nil
}

// CheckProjectPermission 检查用户是否有项目权限
func (s *projectService) CheckProjectPermission(projectID, userID uint, requiredRole string) (bool, error) {
	var permission entity.Permission
	err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// 如果需要特定角色权限
	if requiredRole != "" {
		// 管理员有所有权限
		if permission.Role == entity.ProjectRoleAdmin {
			return true, nil
		}

		// 编辑者有编辑和查看权限
		if permission.Role == entity.ProjectRoleEditor &&
			(requiredRole == entity.ProjectRoleEditor || requiredRole == entity.ProjectRoleViewer) {
			return true, nil
		}

		// 查看者只有查看权限
		if permission.Role == entity.ProjectRoleViewer && requiredRole == entity.ProjectRoleViewer {
			return true, nil
		}

		return false, nil
	}

	// 只检查用户是否是项目成员
	return true, nil
}
