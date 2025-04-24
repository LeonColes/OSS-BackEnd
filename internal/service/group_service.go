package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"oss-backend/internal/model/dto"
	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
)

// GroupService 群组服务接口
type GroupService interface {
	// 群组管理
	CreateGroup(ctx context.Context, req *dto.GroupCreateRequest, creatorID string) error
	UpdateGroup(ctx context.Context, req *dto.GroupUpdateRequest, updaterID string) error
	GetGroupByID(ctx context.Context, id string, userID string) (*dto.GroupResponse, error)
	ListGroups(ctx context.Context, req *dto.GroupListRequest, userID string) (*dto.GroupListResponse, error)

	// 成员管理
	JoinGroup(ctx context.Context, req *dto.GroupJoinRequest, userID string) error
	AddMember(ctx context.Context, groupID string, userID string, role string, operatorID string) error
	UpdateMemberRole(ctx context.Context, groupID string, req *dto.GroupMemberUpdateRequest, operatorID string) error
	RemoveMember(ctx context.Context, groupID string, userID string, operatorID string) error
	ListMembers(ctx context.Context, groupID string, page, size int) (*dto.GroupMemberListResponse, error)

	// 用户群组
	GetUserGroups(ctx context.Context, userID string) ([]dto.GroupResponse, error)
	CheckUserGroupRole(ctx context.Context, groupID string, userID string) (string, error)

	// 邀请码
	GenerateInviteCode(ctx context.Context, req *dto.GroupInviteRequest, userID string) (*dto.GroupInviteResponse, error)
}

// groupService 群组服务实现
type groupService struct {
	groupRepo   repository.GroupRepository
	userRepo    repository.UserRepository
	roleRepo    repository.RoleRepository
	authService AuthService
}

// NewGroupService 创建群组服务
func NewGroupService(
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	authService AuthService,
) GroupService {
	return &groupService{
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		authService: authService,
	}
}

// CreateGroup 创建群组
func (s *groupService) CreateGroup(ctx context.Context, req *dto.GroupCreateRequest, creatorID string) error {
	// 检查群组标识是否已存在
	existingGroup, err := s.groupRepo.GetGroupByKey(ctx, req.GroupKey)
	if err != nil {
		return err
	}
	if existingGroup != nil {
		return fmt.Errorf("群组标识已存在")
	}

	// 生成群组标识
	groupKey := generateGroupKey(req.Name)

	// 生成邀请码
	inviteCode := generateInviteCode()
	expireAt := time.Now().AddDate(0, 0, 30) // 默认30天

	// 创建群组
	group := &entity.Group{
		Name:            req.Name,
		Description:     req.Description,
		GroupKey:        groupKey,
		InviteCode:      inviteCode,
		InviteExpiresAt: &expireAt,
		CreatorID:       creatorID,
		Status:          1, // 正常状态
	}

	err = s.groupRepo.CreateGroup(ctx, group)
	if err != nil {
		return err
	}

	// 将创建者添加为群组成员(管理员)
	member := &entity.GroupMember{
		GroupID:   group.ID,
		UserID:    creatorID,
		Role:      "admin", // 管理员角色
		JoinedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.groupRepo.AddMember(ctx, member)
	if err != nil {
		return err
	}

	// 将创建者添加到Casbin中的GROUP_ADMIN角色
	if s.authService != nil {
		groupDomain := fmt.Sprintf("group:%s", group.ID)
		err = s.authService.AddRoleForUser(ctx, creatorID, entity.RoleGroupAdmin, groupDomain)
		if err != nil {
			// 记录错误但不阻止流程
			// 可以考虑添加日志记录
			fmt.Printf("设置Casbin角色失败: %v\n", err)
		}
	}

	return nil
}

// UpdateGroup 更新群组
func (s *groupService) UpdateGroup(ctx context.Context, req *dto.GroupUpdateRequest, updaterID string) error {
	// 获取群组
	group, err := s.groupRepo.GetGroupByID(ctx, req.ID)
	if err != nil {
		return err
	}

	// 检查用户是否为群组管理员
	role, err := s.CheckUserGroupRole(ctx, req.ID, updaterID)
	if err != nil {
		return err
	}
	if role != "admin" {
		return fmt.Errorf("无权限执行此操作")
	}

	// 更新群组信息
	group.Name = req.Name
	group.Description = req.Description
	if req.Status != nil {
		group.Status = *req.Status
	}

	return s.groupRepo.UpdateGroup(ctx, group)
}

// GetGroupByID 根据ID获取群组
func (s *groupService) GetGroupByID(ctx context.Context, id string, userID string) (*dto.GroupResponse, error) {
	// 获取群组
	group, err := s.groupRepo.GetGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取统计信息
	memberCount, _ := s.groupRepo.GetMemberCount(ctx, id)
	projectCount, _ := s.groupRepo.GetProjectCount(ctx, id)
	storageUsed, _ := s.groupRepo.GetStorageUsed(ctx, id)

	// 获取用户在群组中的角色
	userRole, _ := s.CheckUserGroupRole(ctx, id, userID)

	// 构建响应
	response := &dto.GroupResponse{
		ID:           group.ID,
		Name:         group.Name,
		Description:  group.Description,
		GroupKey:     group.GroupKey,
		StorageQuota: group.StorageQuota,
		StorageUsed:  storageUsed,
		MemberCount:  memberCount,
		ProjectCount: projectCount,
		Status:       group.Status,
		CreatorID:    group.CreatorID,
		CreatedAt:    group.CreatedAt,
		UserRole:     userRole,
	}

	// 添加创建者信息
	if len(group.Creator.ID) > 0 {
		response.CreatorName = group.Creator.Name
	}

	// 只有管理员可以看到邀请码
	if userRole == "admin" {
		response.InviteCode = group.InviteCode
	}

	return response, nil
}

// ListGroups 获取群组列表
func (s *groupService) ListGroups(ctx context.Context, req *dto.GroupListRequest, userID string) (*dto.GroupListResponse, error) {
	// 获取数据
	groups, total, err := s.groupRepo.ListGroups(ctx, req)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &dto.GroupListResponse{
		Total: total,
		Items: make([]dto.GroupResponse, 0, len(groups)),
	}

	for _, group := range groups {
		// 获取统计信息
		memberCount, _ := s.groupRepo.GetMemberCount(ctx, group.ID)
		projectCount, _ := s.groupRepo.GetProjectCount(ctx, group.ID)
		storageUsed, _ := s.groupRepo.GetStorageUsed(ctx, group.ID)

		// 获取用户在群组中的角色
		userRole, _ := s.CheckUserGroupRole(ctx, group.ID, userID)

		item := dto.GroupResponse{
			ID:           group.ID,
			Name:         group.Name,
			Description:  group.Description,
			GroupKey:     group.GroupKey,
			StorageQuota: group.StorageQuota,
			StorageUsed:  storageUsed,
			MemberCount:  memberCount,
			ProjectCount: projectCount,
			Status:       group.Status,
			CreatorID:    group.CreatorID,
			CreatedAt:    group.CreatedAt,
			UserRole:     userRole,
		}

		if len(group.Creator.ID) > 0 {
			item.CreatorName = group.Creator.Name
		}

		// 只有管理员可以看到邀请码
		if userRole == "admin" {
			item.InviteCode = group.InviteCode
		}

		response.Items = append(response.Items, item)
	}

	return response, nil
}

// JoinGroup 加入群组
func (s *groupService) JoinGroup(ctx context.Context, req *dto.GroupJoinRequest, userID string) error {
	// 根据邀请码获取群组
	group, err := s.groupRepo.GetGroupByInviteCode(ctx, req.InviteCode)
	if err != nil {
		return err
	}

	// 检查用户是否已经是群组成员
	member, err := s.groupRepo.GetMember(ctx, group.ID, userID)
	if err != nil {
		return err
	}
	if member != nil {
		return fmt.Errorf("您已经是该群组成员")
	}

	// 添加用户为群组成员
	newMember := &entity.GroupMember{
		GroupID:   group.ID,
		UserID:    userID,
		Role:      "member", // 普通成员角色
		JoinedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.groupRepo.AddMember(ctx, newMember)
}

// AddMember 添加成员
func (s *groupService) AddMember(ctx context.Context, groupID string, userID string, role string, operatorID string) error {
	// 检查群组是否存在
	if _, err := s.groupRepo.GetGroupByID(ctx, groupID); err != nil {
		return err
	}

	// 检查操作者是否为群组管理员
	operatorRole, err := s.CheckUserGroupRole(ctx, groupID, operatorID)
	if err != nil {
		return err
	}
	if operatorRole != "admin" {
		return fmt.Errorf("无权限执行此操作")
	}

	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查用户是否已经是群组成员
	member, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if member != nil {
		return fmt.Errorf("用户已经是该群组成员")
	}

	// 添加用户为群组成员
	newMember := &entity.GroupMember{
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.groupRepo.AddMember(ctx, newMember)
}

// UpdateMemberRole 更新成员角色
func (s *groupService) UpdateMemberRole(ctx context.Context, groupID string, req *dto.GroupMemberUpdateRequest, operatorID string) error {
	// 检查操作者是否为群组管理员
	operatorRole, err := s.CheckUserGroupRole(ctx, groupID, operatorID)
	if err != nil {
		return err
	}
	if operatorRole != "admin" {
		return fmt.Errorf("无权限执行此操作")
	}

	// 不能修改自己的角色
	if req.UserID == operatorID {
		return fmt.Errorf("不能修改自己的角色")
	}

	// 获取成员
	member, err := s.groupRepo.GetMember(ctx, groupID, req.UserID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("用户不是该群组成员")
	}

	// 更新角色
	member.Role = req.Role
	member.UpdatedAt = time.Now()

	return s.groupRepo.UpdateMember(ctx, member)
}

// RemoveMember 移除成员
func (s *groupService) RemoveMember(ctx context.Context, groupID string, userID string, operatorID string) error {
	// 检查操作者是否为群组管理员
	operatorRole, err := s.CheckUserGroupRole(ctx, groupID, operatorID)
	if err != nil {
		return err
	}
	if operatorRole != "admin" {
		return fmt.Errorf("无权限执行此操作")
	}

	// 不能移除自己
	if userID == operatorID {
		return fmt.Errorf("不能移除自己")
	}

	// 检查用户是否为群组成员
	member, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("用户不是该群组成员")
	}

	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

// ListMembers 获取成员列表
func (s *groupService) ListMembers(ctx context.Context, groupID string, page, size int) (*dto.GroupMemberListResponse, error) {
	// 获取数据
	members, total, err := s.groupRepo.ListMembers(ctx, groupID, page, size)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &dto.GroupMemberListResponse{
		Total: total,
		Items: make([]dto.GroupMemberResponse, 0, len(members)),
	}

	for _, member := range members {
		item := dto.GroupMemberResponse{
			ID:           member.ID,
			UserID:       member.UserID,
			Role:         member.Role,
			JoinedAt:     member.JoinedAt,
			LastActiveAt: member.LastActiveAt,
		}

		if len(member.User.ID) > 0 {
			item.UserName = member.User.Name
			item.Email = member.User.Email
			item.Avatar = member.User.Avatar
		}

		response.Items = append(response.Items, item)
	}

	return response, nil
}

// GetUserGroups 获取用户所属的群组
func (s *groupService) GetUserGroups(ctx context.Context, userID string) ([]dto.GroupResponse, error) {
	// 获取数据
	groups, err := s.groupRepo.GetUserGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := make([]dto.GroupResponse, 0, len(groups))

	for _, group := range groups {
		// 获取统计信息
		memberCount, _ := s.groupRepo.GetMemberCount(ctx, group.ID)
		projectCount, _ := s.groupRepo.GetProjectCount(ctx, group.ID)
		storageUsed, _ := s.groupRepo.GetStorageUsed(ctx, group.ID)

		// 获取用户在群组中的角色
		userRole, _ := s.CheckUserGroupRole(ctx, group.ID, userID)

		item := dto.GroupResponse{
			ID:           group.ID,
			Name:         group.Name,
			Description:  group.Description,
			GroupKey:     group.GroupKey,
			StorageQuota: group.StorageQuota,
			StorageUsed:  storageUsed,
			MemberCount:  memberCount,
			ProjectCount: projectCount,
			Status:       group.Status,
			CreatorID:    group.CreatorID,
			CreatedAt:    group.CreatedAt,
			UserRole:     userRole,
		}

		if len(group.Creator.ID) > 0 {
			item.CreatorName = group.Creator.Name
		}

		// 只有管理员可以看到邀请码
		if userRole == "admin" {
			item.InviteCode = group.InviteCode
		}

		response = append(response, item)
	}

	return response, nil
}

// CheckUserGroupRole 检查用户在群组中的角色
func (s *groupService) CheckUserGroupRole(ctx context.Context, groupID string, userID string) (string, error) {
	member, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		return "", err
	}
	if member == nil {
		return "", fmt.Errorf("用户不是该群组成员")
	}
	return member.Role, nil
}

// GenerateInviteCode 生成邀请码
func (s *groupService) GenerateInviteCode(ctx context.Context, req *dto.GroupInviteRequest, userID string) (*dto.GroupInviteResponse, error) {
	// 检查用户是否为群组管理员
	role, err := s.CheckUserGroupRole(ctx, req.GroupID, userID)
	if err != nil {
		return nil, err
	}
	if role != "admin" {
		return nil, fmt.Errorf("无权限执行此操作")
	}

	// 获取群组
	group, err := s.groupRepo.GetGroupByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	// 生成邀请码
	code, expireAt, err := s.groupRepo.GenerateInviteCode(ctx, req.GroupID, req.ExpireDays)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &dto.GroupInviteResponse{
		GroupID:    req.GroupID,
		GroupName:  group.Name,
		InviteCode: code,
		ExpireAt:   &expireAt,
	}

	return response, nil
}

// 生成随机邀请码
func generateInviteCode() string {
	// 简单实现，实际项目中应该使用更复杂的算法
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// 根据群组名生成唯一标识
func generateGroupKey(name string) string {
	// 生成随机字节
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机失败，使用时间戳
		return fmt.Sprintf("%s_%d", name, time.Now().UnixNano())
	}

	// 编码为Base64
	encoded := base64.URLEncoding.EncodeToString(b)
	return fmt.Sprintf("%s_%s", name, encoded[:8])
}
