package test

import (
	"context"
	"errors"
	"testing"

	"oss-backend/mocks"

	"github.com/stretchr/testify/assert"
)

func TestProjectServiceMock_EnsureProjectMemberPermissions(t *testing.T) {
	// 创建 ProjectService 模拟
	mockProjectService := mocks.NewProjectService(t)

	// 测试用例1：成功确保用户权限
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		projectID := "project-123"
		userID := "user-456"

		// 设置期望的调用和返回值
		mockProjectService.On("EnsureProjectMemberPermissions", ctx, projectID, userID).Return(nil).Once()

		// 调用方法
		err := mockProjectService.EnsureProjectMemberPermissions(ctx, projectID, userID)

		// 断言
		assert.NoError(t, err)
		mockProjectService.AssertExpectations(t)
	})

	// 测试用例2：用户不是项目成员
	t.Run("UserNotMember", func(t *testing.T) {
		ctx := context.Background()
		projectID := "project-123"
		userID := "user-789"
		expectedErr := errors.New("用户不是项目成员")

		// 设置期望的调用和返回值
		mockProjectService.On("EnsureProjectMemberPermissions", ctx, projectID, userID).Return(expectedErr).Once()

		// 调用方法
		err := mockProjectService.EnsureProjectMemberPermissions(ctx, projectID, userID)

		// 断言
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockProjectService.AssertExpectations(t)
	})
}
