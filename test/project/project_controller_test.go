package project_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"oss-backend/internal/controller"
	"oss-backend/internal/model/dto"
	"oss-backend/mocks"
	"oss-backend/pkg/common"
)

// 设置测试环境
func setupProjectControllerTest(t *testing.T) (*gin.Context, *httptest.ResponseRecorder, *controller.ProjectController, *mocks.ProjectService) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// 模拟用户已登录
	ctx.Set("userID", "user123")

	// 创建模拟的ProjectService
	mockProjectService := mocks.NewProjectService(t)

	// 创建控制器
	projectController := controller.NewProjectController(mockProjectService)

	return ctx, w, projectController, mockProjectService
}

// TestCreateProject_Success 测试成功创建项目
func TestCreateProject_Success(t *testing.T) {
	// 设置测试环境
	ctx, w, projectController, mockProjectService := setupProjectControllerTest(t)

	// 准备请求数据
	reqData := dto.CreateProjectRequest{
		Name:        "测试项目",
		Description: "这是一个测试项目",
		GroupID:     "group123",
	}

	// 准备模拟响应
	expectedResponse := &dto.ProjectResponse{
		ID:          "project123",
		Name:        "测试项目",
		Description: "这是一个测试项目",
		GroupID:     "group123",
		GroupName:   "测试组",
		CreatorID:   "user123",
		CreatorName: "测试用户",
		Status:      1,
		FileCount:   0,
		TotalSize:   0,
	}

	// 设置模拟服务的行为
	mockProjectService.On("CreateProject", mock.Anything, mock.MatchedBy(func(req *dto.CreateProjectRequest) bool {
		return req.Name == reqData.Name && req.GroupID == reqData.GroupID
	}), "user123").Return(expectedResponse, nil).Once()

	// 准备HTTP请求
	reqBody, _ := json.Marshal(reqData)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/oss/project/create", bytes.NewBuffer(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// 调用控制器方法
	projectController.CreateProject(ctx)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应数据
	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应数据
	assert.Equal(t, common.CodeSuccess, response.Code)
	assert.Equal(t, "success", response.Message)

	// 验证返回的项目数据
	responseData, err := json.Marshal(response.Data)
	assert.NoError(t, err)

	var projectResponse dto.ProjectResponse
	err = json.Unmarshal(responseData, &projectResponse)
	assert.NoError(t, err)

	assert.Equal(t, expectedResponse.ID, projectResponse.ID)
	assert.Equal(t, expectedResponse.Name, projectResponse.Name)
	assert.Equal(t, expectedResponse.GroupID, projectResponse.GroupID)
}

// TestCreateProject_Error 测试创建项目时出错
func TestCreateProject_Error(t *testing.T) {
	// 设置测试环境
	ctx, w, projectController, mockProjectService := setupProjectControllerTest(t)

	// 准备请求数据
	reqData := dto.CreateProjectRequest{
		Name:        "测试项目",
		Description: "这是一个测试项目",
		GroupID:     "group123",
	}

	// 设置模拟服务的行为，返回错误
	mockProjectService.On("CreateProject", mock.Anything, mock.MatchedBy(func(req *dto.CreateProjectRequest) bool {
		return req.Name == reqData.Name && req.GroupID == reqData.GroupID
	}), "user123").Return(nil, errors.New("无权在该分组下创建项目")).Once()

	// 准备HTTP请求
	reqBody, _ := json.Marshal(reqData)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/oss/project/create", bytes.NewBuffer(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// 调用控制器方法
	projectController.CreateProject(ctx)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// 解析响应数据
	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证错误响应
	assert.Equal(t, common.CodeError, response.Code)
	assert.Contains(t, response.Message, "创建项目失败")
	assert.Contains(t, response.Message, "无权在该分组下创建项目")
}

// TestCreateProject_InvalidRequest 测试无效请求
func TestCreateProject_InvalidRequest(t *testing.T) {
	// 设置测试环境
	ctx, w, projectController, _ := setupProjectControllerTest(t)

	// 准备无效的请求数据（缺少必要字段）
	reqData := map[string]interface{}{
		"description": "这是一个测试项目",
		// 缺少name和group_id字段
	}

	// 准备HTTP请求
	reqBody, _ := json.Marshal(reqData)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/oss/project/create", bytes.NewBuffer(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// 调用控制器方法
	projectController.CreateProject(ctx)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 解析响应数据
	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证错误响应
	assert.Equal(t, common.CodeError, response.Code)
	assert.Contains(t, response.Message, "请求参数错误")
}
