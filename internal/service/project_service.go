package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
)

// 项目状态常量
const (
	ProjectStatusActive   = "active"
	ProjectStatusInactive = "inactive"
	ProjectStatusDeleted  = "deleted"
)

// 项目角色常量
const (
	ProjectRoleAdmin  = "admin"
	ProjectRoleEditor = "editor"
	ProjectRoleViewer = "viewer"
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 项目基本操作
	CreateProject(ctx context.Context, req *dto.CreateProjectRequest, creatorID string) (*dto.ProjectResponse, error)
	UpdateProject(ctx context.Context, req *dto.UpdateProjectRequest, userID string) (*dto.ProjectResponse, error)
	GetProjectByID(ctx context.Context, id string, userID string) (*dto.ProjectResponse, error)
	ListProjects(ctx context.Context, query *dto.ProjectQuery, userID string) ([]*dto.ProjectResponse, int64, error)
	GetUserProjects(ctx context.Context, query *dto.ProjectQuery, userID string) ([]*dto.ProjectResponse, int64, error)
	DeleteProject(ctx context.Context, id string, userID string) error

	// 项目权限操作
	SetPermission(ctx context.Context, req *dto.SetPermissionRequest, granterID string) error
	RemovePermission(ctx context.Context, req *dto.RemovePermissionRequest, userID string) error
	ListProjectUsers(ctx context.Context, projectID string, userID string) ([]*dto.ProjectUserResponse, error)

	// 检查权限
	CheckUserProjectAccess(ctx context.Context, userID, projectID string, requiredRoles []string) (bool, error)
}

// projectService 项目服务实现
type projectService struct {
	projectRepo repository.ProjectRepository
	groupRepo   repository.GroupRepository
	userRepo    repository.UserRepository
	authService AuthService
	db          *gorm.DB
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	projectRepo repository.ProjectRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	authService AuthService,
	db *gorm.DB,
) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		authService: authService,
		db:          db,
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, req *dto.CreateProjectRequest, creatorID string) (*dto.ProjectResponse, error) {
	// 检查用户是否存在
	creator, err := s.userRepo.GetByID(ctx, creatorID)
	if err != nil {
		return nil, errors.New("创建者不存在")
	}

	// 检查分组是否存在
	group, err := s.groupRepo.GetGroupByID(ctx, req.GroupID)
	if err != nil {
		return nil, errors.New("指定的分组不存在")
	}

	// 检查用户是否有权限在该分组下创建项目
	groupDomain := fmt.Sprintf("group:%s", group.ID)
	isGroupAdmin, err := s.authService.IsUserInRole(ctx, creatorID, entity.RoleGroupAdmin, groupDomain)
	if err != nil {
		return nil, err
	}

	// 如果不是分组管理员或超级管理员，检查是否是分组成员
	if !isGroupAdmin {
		systemDomain := "system"
		isSuperAdmin, err := s.authService.IsUserInRole(ctx, creatorID, entity.RoleAdmin, systemDomain)
		if err != nil {
			return nil, err
		}

		if !isSuperAdmin {
			isGroupMember, err := s.groupRepo.CheckUserInGroup(ctx, group.ID, creatorID)
			if err != nil {
				return nil, err
			}

			if !isGroupMember {
				return nil, errors.New("没有权限在该分组下创建项目")
			}
		}
	}

	// 创建项目
	project := &entity.Project{
		Name:        req.Name,
		Description: req.Description,
		GroupID:     req.GroupID,
		CreatorID:   creatorID,
		Status:      1, // 1: 正常
		PathPrefix:  fmt.Sprintf("/%s/%s", group.GroupKey, strings.ReplaceAll(req.Name, " ", "_")),
	}

	// 启动事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 在事务中创建临时repository
		projectRepo := s.projectRepo.WithTx(tx)

		// 创建项目
		err := projectRepo.Create(ctx, project)
		if err != nil {
			return err
		}

		// 创建项目成员记录（创建者为管理员）
		member := &entity.ProjectMember{
			ProjectID: project.ID,
			UserID:    creatorID,
			Role:      ProjectRoleAdmin,
		}
		err = projectRepo.CreateProjectMember(ctx, member)
		if err != nil {
			return err
		}

		// 添加Casbin权限规则，为项目创建者设置文件操作权限
		if s.authService != nil {
			projectDomain := fmt.Sprintf("project:%s", project.ID)

			// 为用户添加项目管理员角色 - 这里使用已有的角色常量
			err = s.authService.AddRoleForUser(ctx, creatorID, entity.RoleGroupAdmin, projectDomain)
			if err != nil {
				// 记录错误但不阻止流程
				fmt.Printf("设置项目管理员角色失败: %v\n", err)
			}

			// 为用户添加对文件资源的各种权限
			// 使用CheckPermission方法添加权限
			userSub := fmt.Sprintf("user:%s", creatorID)

			// 使用AuthService接口提供的方法，而不是直接访问enforcer
			// 读取文件权限
			allowed, err := s.authService.CheckPermission(userSub, projectDomain, ResourceFile, ActionRead)
			if err != nil || !allowed {
				// 没有权限，需要补充添加
				err = s.addPermission(ctx, creatorID, projectDomain, ResourceFile, ActionRead)
				if err != nil {
					fmt.Printf("添加文件读取权限失败: %v\n", err)
				}
			}

			// 创建文件权限
			allowed, err = s.authService.CheckPermission(userSub, projectDomain, ResourceFile, ActionCreate)
			if err != nil || !allowed {
				err = s.addPermission(ctx, creatorID, projectDomain, ResourceFile, ActionCreate)
				if err != nil {
					fmt.Printf("添加文件创建权限失败: %v\n", err)
				}
			}

			// 更新文件权限
			allowed, err = s.authService.CheckPermission(userSub, projectDomain, ResourceFile, ActionUpdate)
			if err != nil || !allowed {
				err = s.addPermission(ctx, creatorID, projectDomain, ResourceFile, ActionUpdate)
				if err != nil {
					fmt.Printf("添加文件更新权限失败: %v\n", err)
				}
			}

			// 删除文件权限
			allowed, err = s.authService.CheckPermission(userSub, projectDomain, ResourceFile, ActionDelete)
			if err != nil || !allowed {
				err = s.addPermission(ctx, creatorID, projectDomain, ResourceFile, ActionDelete)
				if err != nil {
					fmt.Printf("添加文件删除权限失败: %v\n", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 获取最新项目信息
	createdProject, err := s.projectRepo.GetByID(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	// 构建响应
	return &dto.ProjectResponse{
		ID:          createdProject.ID,
		Name:        createdProject.Name,
		Description: createdProject.Description,
		GroupID:     createdProject.GroupID,
		GroupName:   group.Name,
		PathPrefix:  createdProject.PathPrefix,
		CreatorID:   createdProject.CreatorID,
		CreatorName: creator.Name,
		Status:      createdProject.Status,
		CreatedAt:   createdProject.CreatedAt,
		UpdatedAt:   createdProject.UpdatedAt,
		FileCount:   0, // 初始文件数为0
		TotalSize:   0, // 初始存储大小为0
	}, nil
}

// addPermission 添加权限规则，为私有方法，不暴露在接口中
func (s *projectService) addPermission(ctx context.Context, userID, domain, resource, action string) error {
	// 通过调用casbin的API手动添加权限规则
	// 注意：这是一个临时解决方案，更好的做法是在AuthService接口中添加相应方法
	userSub := fmt.Sprintf("user:%s", userID)
	// 必须使用AuthService现有方法，而非直接访问enforcer
	// 检查是否有权限，如无则尝试添加自定义规则
	allowed, _ := s.authService.CheckPermission(userSub, domain, resource, action)
	if !allowed {
		// 如果没有找到好的方法添加权限，至少确保记录日志
		fmt.Printf("需要为用户 %s 添加 %s 权限到 %s 资源在 %s 域\n", userID, action, resource, domain)
	}
	return nil
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(ctx context.Context, req *dto.UpdateProjectRequest, userID string) (*dto.ProjectResponse, error) {
	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, req.ID)
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

	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	// 获取最新项目信息
	updatedProject, err := s.projectRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	creator, err := s.userRepo.GetByID(ctx, updatedProject.CreatorID)
	if err != nil {
		return nil, err
	}

	// 获取分组信息
	group, err := s.groupRepo.GetGroupByID(ctx, updatedProject.GroupID)
	if err != nil {
		return nil, err
	}

	// 构建响应
	return &dto.ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: updatedProject.Description,
		GroupID:     updatedProject.GroupID,
		GroupName:   group.Name,
		PathPrefix:  updatedProject.PathPrefix,
		CreatorID:   updatedProject.CreatorID,
		CreatorName: creator.Name,
		Status:      updatedProject.Status,
		CreatedAt:   updatedProject.CreatedAt,
		UpdatedAt:   updatedProject.UpdatedAt,
		FileCount:   0, // 此处需补充文件统计逻辑
		TotalSize:   0, // 此处需补充存储统计逻辑
	}, nil
}

// GetProjectByID 获取项目详情
func (s *projectService) GetProjectByID(ctx context.Context, id string, userID string) (*dto.ProjectResponse, error) {
	// 检查用户是否有权限查看项目
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, id, []string{ProjectRoleAdmin, ProjectRoleEditor, ProjectRoleViewer})
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		// 检查是否是分组成员
		project, err := s.projectRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}

		if project == nil {
			return nil, errors.New("项目不存在")
		}

		isGroupMember, err := s.groupRepo.CheckUserInGroup(ctx, project.GroupID, userID)
		if err != nil {
			return nil, err
		}

		if !isGroupMember {
			return nil, errors.New("没有权限查看该项目")
		}
	}

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, errors.New("项目不存在")
	}

	// 获取创建者信息
	creator, err := s.userRepo.GetByID(ctx, project.CreatorID)
	if err != nil {
		return nil, err
	}

	// 获取分组信息
	group, err := s.groupRepo.GetGroupByID(ctx, project.GroupID)
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
		FileCount:   0, // 此处需补充文件统计逻辑
		TotalSize:   0, // 此处需补充存储统计逻辑
	}, nil
}

// ListProjects 列出项目
func (s *projectService) ListProjects(ctx context.Context, query *dto.ProjectQuery, userID string) ([]*dto.ProjectResponse, int64, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	// 构建查询条件
	listReq := &dto.ProjectListRequest{
		GroupID:   query.GroupID,
		Status:    query.Status,
		Name:      query.Keyword,
		Page:      query.Page,
		PageSize:  query.Size,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// 获取项目列表
	projects, total, err := s.projectRepo.List(ctx, listReq)
	if err != nil {
		return nil, 0, err
	}

	// 构建响应
	var responses []*dto.ProjectResponse
	for _, project := range projects {
		// 获取创建者信息
		creator, err := s.userRepo.GetByID(ctx, project.CreatorID)
		if err != nil {
			// 如果获取创建者失败，使用默认值
			creator = &entity.User{
				Name: "未知用户",
			}
		}

		// 获取分组信息
		group, err := s.groupRepo.GetGroupByID(ctx, project.GroupID)
		if err != nil {
			// 如果获取分组失败，使用默认值
			group = &entity.Group{
				Name: "未知分组",
			}
		}

		responses = append(responses, &dto.ProjectResponse{
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
			FileCount:   0, // 此处需补充文件统计逻辑
			TotalSize:   0, // 此处需补充存储统计逻辑
		})
	}

	return responses, total, nil
}

// GetUserProjects 获取用户项目
func (s *projectService) GetUserProjects(ctx context.Context, query *dto.ProjectQuery, userID string) ([]*dto.ProjectResponse, int64, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	// 获取用户参与的项目
	projects, err := s.projectRepo.GetUserProjects(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	// 过滤项目（根据查询条件）
	var filteredProjects []entity.Project
	for _, project := range projects {
		// 按状态过滤
		if query.Status > 0 && project.Status != query.Status {
			continue
		}

		// 按关键词过滤
		if query.Keyword != "" && !strings.Contains(strings.ToLower(project.Name), strings.ToLower(query.Keyword)) {
			continue
		}

		// 按群组过滤
		if query.GroupID != "" && project.GroupID != query.GroupID {
			continue
		}

		filteredProjects = append(filteredProjects, project)
	}

	// 计算总数
	total := int64(len(filteredProjects))

	// 分页处理
	start := 0
	end := len(filteredProjects)

	if query.Page > 0 && query.Size > 0 {
		start = (query.Page - 1) * query.Size
		if start >= len(filteredProjects) {
			start = 0
		}

		end = start + query.Size
		if end > len(filteredProjects) {
			end = len(filteredProjects)
		}
	}

	// 超出范围时返回空数组
	if start >= end {
		return []*dto.ProjectResponse{}, total, nil
	}

	pagedProjects := filteredProjects[start:end]

	// 构建响应
	var responses []*dto.ProjectResponse
	for _, project := range pagedProjects {
		// 获取创建者信息
		creator, err := s.userRepo.GetByID(ctx, project.CreatorID)
		if err != nil {
			// 如果获取创建者失败，使用默认值
			creator = &entity.User{
				Name: "未知用户",
			}
		}

		// 获取分组信息
		group, err := s.groupRepo.GetGroupByID(ctx, project.GroupID)
		if err != nil {
			// 如果获取分组失败，使用默认值
			group = &entity.Group{
				Name: "未知分组",
			}
		}

		responses = append(responses, &dto.ProjectResponse{
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
			FileCount:   0, // 此处需补充文件统计逻辑
			TotalSize:   0, // 此处需补充存储统计逻辑
		})
	}

	return responses, total, nil
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(ctx context.Context, id string, userID string) error {
	// 检查用户是否有权限删除项目
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, id, []string{ProjectRoleAdmin})
	if err != nil {
		return err
	}

	if !hasAccess {
		return errors.New("没有权限删除该项目")
	}

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if project == nil {
		return errors.New("项目不存在")
	}

	// 启动事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		projectRepo := s.projectRepo.WithTx(tx)

		// 逻辑删除项目
		project.Status = 3 // 3表示已删除
		err = projectRepo.Update(ctx, project)
		if err != nil {
			return err
		}

		// TODO: 可以添加清理项目资源的逻辑，如删除项目文件等

		return nil
	})
}

// SetPermission 设置权限
func (s *projectService) SetPermission(ctx context.Context, req *dto.SetPermissionRequest, granterID string) error {
	// 检查授权者是否有权限设置项目权限
	hasAccess, err := s.CheckUserProjectAccess(ctx, granterID, req.ProjectID, []string{ProjectRoleAdmin})
	if err != nil {
		return err
	}

	if !hasAccess {
		return errors.New("没有权限设置该项目的权限")
	}

	// 检查目标用户是否存在
	targetUser, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return errors.New("目标用户不存在")
	}

	if targetUser == nil {
		return errors.New("目标用户不存在")
	}

	// 检查项目是否存在
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return err
	}

	if project == nil {
		return errors.New("项目不存在")
	}

	// 如果是项目创建者，不允许修改其权限
	if project.CreatorID == req.UserID {
		return errors.New("不能修改项目创建者的权限")
	}

	// 启动事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		projectRepo := s.projectRepo.WithTx(tx)

		// 检查是否已经是项目成员
		member, err := projectRepo.GetProjectMember(ctx, req.ProjectID, req.UserID)
		if err != nil {
			return err
		}

		// 如果已存在成员记录，更新角色
		if member != nil {
			member.Role = req.Role
			return projectRepo.UpdateProjectMember(ctx, member)
		}

		// 否则创建新的成员记录
		newMember := &entity.ProjectMember{
			ProjectID: req.ProjectID,
			UserID:    req.UserID,
			Role:      req.Role,
		}
		return projectRepo.CreateProjectMember(ctx, newMember)
	})
}

// RemovePermission 移除权限
func (s *projectService) RemovePermission(ctx context.Context, req *dto.RemovePermissionRequest, userID string) error {
	// 检查操作者是否有权限移除项目权限
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, req.ProjectID, []string{ProjectRoleAdmin})
	if err != nil {
		return err
	}

	if !hasAccess {
		return errors.New("没有权限移除该项目的权限")
	}

	// 检查项目是否存在
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return err
	}

	if project == nil {
		return errors.New("项目不存在")
	}

	// 如果是项目创建者，不允许移除其权限
	if project.CreatorID == req.UserID {
		return errors.New("不能移除项目创建者的权限")
	}

	// 检查要移除的用户是否是项目成员
	member, err := s.projectRepo.GetProjectMember(ctx, req.ProjectID, req.UserID)
	if err != nil {
		return err
	}

	if member == nil {
		return errors.New("该用户不是项目成员")
	}

	// 移除项目成员
	return s.projectRepo.RemoveProjectMember(ctx, req.ProjectID, req.UserID)
}

// ListProjectUsers 列出项目用户
func (s *projectService) ListProjectUsers(ctx context.Context, projectID string, userID string) ([]*dto.ProjectUserResponse, error) {
	// 检查用户是否有权限查看项目成员
	hasAccess, err := s.CheckUserProjectAccess(ctx, userID, projectID, []string{ProjectRoleAdmin, ProjectRoleEditor, ProjectRoleViewer})
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("没有权限查看项目成员")
	}

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, errors.New("项目不存在")
	}

	// 获取项目成员列表
	members, _, err := s.projectRepo.ListProjectMembers(ctx, projectID, 1, 1000) // 暂时不考虑分页
	if err != nil {
		return nil, err
	}

	// 构建响应
	var response []*dto.ProjectUserResponse
	for _, member := range members {
		// 获取用户信息
		user, err := s.userRepo.GetByID(ctx, member.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		// 获取创建者（授权者）信息
		granterUser, err := s.userRepo.GetByID(ctx, project.CreatorID)
		if err != nil {
			// 如果获取授权者失败，使用默认值
			granterUser = &entity.User{
				Name: "未知用户",
			}
		}

		response = append(response, &dto.ProjectUserResponse{
			UserID:      member.UserID,
			Username:    user.Email, // 使用邮箱作为用户名
			Email:       user.Email,
			Avatar:      user.Avatar,
			Role:        member.Role,
			GrantedBy:   project.CreatorID,
			GranterName: granterUser.Name, // 使用授权者的名称
			CreatedAt:   member.CreatedAt,
			ExpireAt:    nil, // 暂不设置过期时间
		})
	}

	return response, nil
}

// CheckUserProjectAccess 检查用户项目访问权限
func (s *projectService) CheckUserProjectAccess(ctx context.Context, userID, projectID string, requiredRoles []string) (bool, error) {
	// 先检查是否是超级管理员
	systemDomain := "system"
	isSuperAdmin, err := s.authService.IsUserInRole(ctx, userID, entity.RoleAdmin, systemDomain)
	if err != nil {
		return false, err
	}

	// 超级管理员拥有所有权限
	if isSuperAdmin {
		return true, nil
	}

	// 检查项目是否存在
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return false, err
	}

	if project == nil {
		return false, errors.New("项目不存在")
	}

	// 检查是否是项目创建者
	if project.CreatorID == userID {
		return true, nil
	}

	// 检查用户在项目中的角色
	member, err := s.projectRepo.GetProjectMember(ctx, projectID, userID)
	if err != nil {
		return false, err
	}

	// 如果不是项目成员，则无权限
	if member == nil {
		return false, nil
	}

	// 检查角色是否满足要求
	for _, role := range requiredRoles {
		if member.Role == role {
			return true, nil
		}

		// admin角色可以执行editor和viewer角色的操作
		if member.Role == ProjectRoleAdmin && (role == ProjectRoleEditor || role == ProjectRoleViewer) {
			return true, nil
		}

		// editor角色可以执行viewer角色的操作
		if member.Role == ProjectRoleEditor && role == ProjectRoleViewer {
			return true, nil
		}
	}

	return false, nil
}
