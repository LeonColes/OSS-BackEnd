package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
)

// 项目状态常量
const (
	ProjectStatusNormal  = 1 // 正常
	ProjectStatusArchive = 2 // 归档
	ProjectStatusDeleted = 3 // 删除
)

// 项目角色常量
const (
	ProjectRoleAdmin  = "admin"  // 管理员
	ProjectRoleEditor = "editor" // 编辑者
	ProjectRoleViewer = "viewer" // 查看者
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 项目基本操作
	CreateProject(ctx context.Context, req *dto.CreateProjectRequest, creatorID uint64) (*dto.ProjectResponse, error)
	UpdateProject(ctx context.Context, req *dto.UpdateProjectRequest, userID uint64) (*dto.ProjectResponse, error)
	GetProjectByID(ctx context.Context, id uint64, userID uint64) (*dto.ProjectResponse, error)
	ListProjects(ctx context.Context, query *dto.ProjectQuery, userID uint64) ([]*dto.ProjectResponse, int64, error)
	GetUserProjects(ctx context.Context, query *dto.ProjectQuery, userID uint64) ([]*dto.ProjectResponse, int64, error)
	DeleteProject(ctx context.Context, id uint64, userID uint64) error

	// 项目权限操作
	SetPermission(ctx context.Context, req *dto.SetPermissionRequest, granterID uint64) error
	RemovePermission(ctx context.Context, req *dto.RemovePermissionRequest, userID uint64) error
	ListProjectUsers(ctx context.Context, projectID uint64, userID uint64) ([]*dto.ProjectUserResponse, error)

	// 检查权限
	CheckUserProjectAccess(ctx context.Context, userID, projectID uint64, requiredRoles []string) (bool, error)
}

// projectService 项目服务实现
type projectService struct {
	projectRepo repository.ProjectRepository
	groupRepo   repository.GroupRepository
	userRepo    repository.UserRepository
	authSvc     AuthService
}

// NewProjectService 创建项目服务
func NewProjectService(
	projectRepo repository.ProjectRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	authSvc AuthService,
) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		authSvc:     authSvc,
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, req *dto.CreateProjectRequest, creatorID uint64) (*dto.ProjectResponse, error) {
	// 检查群组是否存在
	group, err := s.groupRepo.GetGroupByID(ctx, req.GroupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}

	// 检查用户是否是群组管理员
	isAdmin, err := s.groupRepo.CheckUserGroupRole(ctx, creatorID, req.GroupID, "admin")
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		return nil, errors.New("只有群组管理员才能创建项目")
	}

	// 创建项目路径前缀（群组ID+项目名）
	pathPrefix := fmt.Sprintf("/%d/%s", req.GroupID, strings.ReplaceAll(req.Name, " ", "_"))

	// 创建项目
	project := &entity.Project{
		GroupID:     req.GroupID,
		Name:        req.Name,
		Description: req.Description,
		PathPrefix:  pathPrefix,
		CreatorID:   creatorID,
		Status:      ProjectStatusNormal,
	}

	err = s.projectRepo.CreateProject(ctx, project)
	if err != nil {
		return nil, err
	}

	// 为创建者添加项目管理员权限
	permission := &entity.Permission{
		UserID:    creatorID,
		ProjectID: project.ID,
		Role:      ProjectRoleAdmin,
		GrantedBy: creatorID,
	}

	err = s.projectRepo.SetPermission(ctx, permission)
	if err != nil {
		return nil, err
	}

	// 添加项目管理员权限到Casbin
	domain := fmt.Sprintf("project:%d", project.ID)
	err = s.authSvc.AddRoleForUser(ctx, creatorID, ProjectRoleAdmin, domain)
	if err != nil {
		return nil, err
	}

	// 获取创建者信息
	creator, err := s.userRepo.GetByID(ctx, creatorID)
	if err != nil {
		return nil, err
	}

	// 构建响应
	return &dto.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		GroupID:     project.GroupID,
		GroupName:   group.Name,
		PathPrefix:  project.PathPrefix,
		CreatorID:   project.CreatorID,
		CreatorName: creator.Name,
		Status:      project.Status,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		FileCount:   0,
		TotalSize:   0,
	}, nil
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(ctx context.Context, req *dto.UpdateProjectRequest, userID uint64) (*dto.ProjectResponse, error) {
	// 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// 检查用户是否有权限更新项目
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, req.ID, []string{ProjectRoleAdmin})
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("没有权限更新项目信息")
	}

	// 更新项目信息
	project.Name = req.Name
	project.Description = req.Description
	if req.Status > 0 {
		project.Status = req.Status
	}

	err = s.projectRepo.UpdateProject(ctx, project)
	if err != nil {
		return nil, err
	}

	// 获取最新项目信息
	updatedProject, err := s.projectRepo.GetProjectByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 构建响应
	return &dto.ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: updatedProject.Description,
		GroupID:     updatedProject.GroupID,
		GroupName:   updatedProject.Group.Name,
		PathPrefix:  updatedProject.PathPrefix,
		CreatorID:   updatedProject.CreatorID,
		CreatorName: updatedProject.Creator.Name,
		Status:      updatedProject.Status,
		CreatedAt:   updatedProject.CreatedAt,
		UpdatedAt:   updatedProject.UpdatedAt,
		FileCount:   0, // 此处需补充文件统计逻辑
		TotalSize:   0, // 此处需补充存储统计逻辑
	}, nil
}

// GetProjectByID 获取项目详情
func (s *projectService) GetProjectByID(ctx context.Context, id uint64, userID uint64) (*dto.ProjectResponse, error) {
	// 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// 检查用户是否有权限查看项目
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, id, []string{ProjectRoleAdmin, ProjectRoleEditor, ProjectRoleViewer})
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("没有权限查看此项目")
	}

	// 构建响应
	return &dto.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		GroupID:     project.GroupID,
		GroupName:   project.Group.Name,
		PathPrefix:  project.PathPrefix,
		CreatorID:   project.CreatorID,
		CreatorName: project.Creator.Name,
		Status:      project.Status,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		FileCount:   0, // 此处需补充文件统计逻辑
		TotalSize:   0, // 此处需补充存储统计逻辑
	}, nil
}

// ListProjects 列出项目
func (s *projectService) ListProjects(ctx context.Context, query *dto.ProjectQuery, userID uint64) ([]*dto.ProjectResponse, int64, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	// 获取项目列表
	projects, total, err := s.projectRepo.ListProjects(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 构建响应
	var responses []*dto.ProjectResponse
	for _, project := range projects {
		responses = append(responses, &dto.ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			GroupID:     project.GroupID,
			GroupName:   project.Group.Name,
			PathPrefix:  project.PathPrefix,
			CreatorID:   project.CreatorID,
			CreatorName: project.Creator.Name,
			Status:      project.Status,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
			FileCount:   0, // 此处需补充文件统计逻辑
			TotalSize:   0, // 此处需补充存储统计逻辑
		})
	}

	return responses, total, nil
}

// GetUserProjects 获取用户参与的项目
func (s *projectService) GetUserProjects(ctx context.Context, query *dto.ProjectQuery, userID uint64) ([]*dto.ProjectResponse, int64, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	// 获取用户参与的项目
	projects, total, err := s.projectRepo.GetUserProjects(ctx, userID, query)
	if err != nil {
		return nil, 0, err
	}

	// 构建响应
	var responses []*dto.ProjectResponse
	for _, project := range projects {
		responses = append(responses, &dto.ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			GroupID:     project.GroupID,
			GroupName:   project.Group.Name,
			PathPrefix:  project.PathPrefix,
			CreatorID:   project.CreatorID,
			CreatorName: project.Creator.Name,
			Status:      project.Status,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
			FileCount:   0, // 此处需补充文件统计逻辑
			TotalSize:   0, // 此处需补充存储统计逻辑
		})
	}

	return responses, total, nil
}

// DeleteProject 删除项目（逻辑删除）
func (s *projectService) DeleteProject(ctx context.Context, id uint64, userID uint64) error {
	// 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}

	// 检查用户是否有权限删除项目
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, id, []string{ProjectRoleAdmin})
	if err != nil {
		return err
	}

	if !hasAccess {
		// 检查用户是否是群组管理员
		isGroupAdmin, err := s.groupRepo.CheckUserGroupRole(ctx, userID, project.GroupID, "admin")
		if err != nil {
			return err
		}

		if !isGroupAdmin {
			return errors.New("没有权限删除此项目")
		}
	}

	// 执行删除操作（逻辑删除）
	return s.projectRepo.DeleteProject(ctx, id)
}

// SetPermission 设置项目权限
func (s *projectService) SetPermission(ctx context.Context, req *dto.SetPermissionRequest, granterID uint64) error {
	// 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}

	// 检查用户是否有权限设置项目权限
	hasAccess, err := s.CheckUserProjectAccess(ctx, granterID, req.ProjectID, []string{ProjectRoleAdmin})
	if err != nil {
		return err
	}

	if !hasAccess {
		// 检查用户是否是群组管理员
		isGroupAdmin, err := s.groupRepo.CheckUserGroupRole(ctx, granterID, project.GroupID, "admin")
		if err != nil {
			return err
		}

		if !isGroupAdmin {
			return errors.New("没有权限管理项目成员")
		}
	}

	// 检查目标用户是否存在
	_, err = s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("目标用户不存在")
		}
		return err
	}

	// 检查用户是否属于同一个群组
	isMember, err := s.groupRepo.CheckUserInGroup(ctx, req.UserID, project.GroupID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("目标用户不在此项目所属群组中")
	}

	// 解析过期时间
	var expireAt *time.Time
	if req.ExpireAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.ExpireAt)
		if err != nil {
			return errors.New("日期格式错误，请使用YYYY-MM-DD HH:MM:SS")
		}
		expireAt = &t
	}

	// 创建或更新权限
	permission := &entity.Permission{
		UserID:    req.UserID,
		ProjectID: req.ProjectID,
		Role:      req.Role,
		GrantedBy: granterID,
		ExpireAt:  expireAt,
	}

	err = s.projectRepo.SetPermission(ctx, permission)
	if err != nil {
		return err
	}

	// 更新Casbin权限
	domain := fmt.Sprintf("project:%d", req.ProjectID)
	// sub := fmt.Sprintf("user:%d", req.UserID)

	// 先移除旧角色
	subject := fmt.Sprintf("user:%d", req.UserID)
	roles, err := s.authSvc.GetRolesForUser(subject, domain)
	if err != nil {
		return err
	}

	for _, role := range roles {
		err = s.authSvc.RemoveRoleForUser(ctx, req.UserID, role, domain)
		if err != nil {
			return err
		}
	}

	// 添加新角色
	err = s.authSvc.AddRoleForUser(ctx, req.UserID, req.Role, domain)
	if err != nil {
		return err
	}

	return nil
}

// RemovePermission 移除项目权限
func (s *projectService) RemovePermission(ctx context.Context, req *dto.RemovePermissionRequest, userID uint64) error {
	// 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}

	// 检查用户是否有权限移除项目权限
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, req.ProjectID, []string{ProjectRoleAdmin})
	if err != nil {
		return err
	}

	if !hasAccess {
		// 检查用户是否是群组管理员
		isGroupAdmin, err := s.groupRepo.CheckUserGroupRole(ctx, userID, project.GroupID, "admin")
		if err != nil {
			return err
		}

		if !isGroupAdmin {
			return errors.New("没有权限管理项目成员")
		}
	}

	// 检查目标用户是否存在
	_, err = s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("目标用户不存在")
		}
		return err
	}

	// 检查是否是项目创建者
	if project.CreatorID == req.UserID {
		return errors.New("不能移除项目创建者的权限")
	}

	// 移除权限
	err = s.projectRepo.RemovePermission(ctx, req.ProjectID, req.UserID)
	if err != nil {
		return err
	}

	// 更新Casbin权限
	domain := fmt.Sprintf("project:%d", req.ProjectID)
	// sub := fmt.Sprintf("user:%d", req.UserID)

	// 移除角色
	subject := fmt.Sprintf("user:%d", req.UserID)
	roles, err := s.authSvc.GetRolesForUser(subject, domain)
	if err != nil {
		return err
	}

	for _, role := range roles {
		err = s.authSvc.RemoveRoleForUser(ctx, req.UserID, role, domain)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListProjectUsers 列出项目用户
func (s *projectService) ListProjectUsers(ctx context.Context, projectID uint64, userID uint64) ([]*dto.ProjectUserResponse, error) {
	// 获取项目信息
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// 检查用户是否有权限查看项目成员
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, projectID, []string{ProjectRoleAdmin, ProjectRoleEditor, ProjectRoleViewer})
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		// 检查用户是否是群组管理员
		isGroupAdmin, err := s.groupRepo.CheckUserGroupRole(ctx, userID, project.GroupID, "admin")
		if err != nil {
			return nil, err
		}

		if !isGroupAdmin {
			return nil, errors.New("没有权限查看项目成员")
		}
	}

	// 获取项目用户列表
	permissions, err := s.projectRepo.ListProjectUsers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 构建响应
	var responses []*dto.ProjectUserResponse
	for _, permission := range permissions {
		resp := &dto.ProjectUserResponse{
			UserID:      permission.UserID,
			Username:    permission.User.Name,
			Email:       permission.User.Email,
			Avatar:      permission.User.Avatar,
			Role:        permission.Role,
			GrantedBy:   permission.GrantedBy,
			GranterName: permission.Granter.Name,
			CreatedAt:   permission.CreatedAt,
			ExpireAt:    permission.ExpireAt,
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// CheckUserProjectAccess 检查用户是否有项目权限
func (s *projectService) CheckUserProjectAccess(ctx context.Context, userID, projectID uint64, requiredRoles []string) (bool, error) {
	return s.projectRepo.CheckUserProjectRole(ctx, userID, projectID, requiredRoles)
}
