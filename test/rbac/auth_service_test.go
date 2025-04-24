package rbac_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/stretchr/testify/assert" // Needed for mock args

	"oss-backend/internal/model/entity"
	"oss-backend/internal/service"

	// Import mocks directly from the root mocks directory
	mocks "oss-backend/mocks"
)

// setupAuthServiceTest creates mocks and initializes the authService with memory enforcer and nil DB
// Returns the AuthService interface, repository mocks, and the enforcer.
func setupAuthServiceTest(t *testing.T) (service.AuthService, *mocks.RoleRepository, *mocks.UserRepository, *mocks.CasbinRepository, *casbin.Enforcer) {
	mockRoleRepo := mocks.NewRoleRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)
	mockCasbinRepo := mocks.NewCasbinRepository(t)

	// --- Initialize Casbin Enforcer with Memory Adapter ---
	modelText := `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _
g2 = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act || g2(r.sub, p.sub) && p.dom == "system" && p.obj == "all" && p.act == "manage"
`
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		log.Fatalf("Failed to create casbin model from string: %v", err)
	}
	adapter := fileadapter.NewAdapter("") // Empty policy file path means memory adapter
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Fatalf("Failed to create casbin enforcer: %v", err)
	}

	// --- Gorm DB Initialization Removed ---

	// --- Create the service instance ---
	authSvcInstance := service.NewAuthService(
		enforcer, // Pass the real enforcer
		mockRoleRepo,
		mockUserRepo,
		mockCasbinRepo,
		nil, // Pass nil for *gorm.DB
	)

	return authSvcInstance, mockRoleRepo, mockUserRepo, mockCasbinRepo, enforcer
}

func TestAuthService_AssignRolesToUser(t *testing.T) {
	authSvc, mockRoleRepo, mockUserRepo, _, enforcer := setupAuthServiceTest(t)
	ctx := context.Background()
	userID := "user1"
	roleIDs := []uint{1, 2}
	domain := "project:proj1"
	roles := []entity.Role{
		{ID: 1, Code: "EDITOR"},
		{ID: 2, Code: "VIEWER"},
	}
	userSub := fmt.Sprintf("user:%s", userID)

	// Mock repository interactions
	mockRoleRepo.On("GetByID", ctx, uint(1)).Return(&roles[0], nil).Once()
	mockRoleRepo.On("GetByID", ctx, uint(2)).Return(&roles[1], nil).Once()
	mockUserRepo.On("AssignRoles", ctx, userID, roleIDs).Return(nil).Once()

	// Call the service method
	err := authSvc.AssignRolesToUser(ctx, userID, roleIDs, domain)

	// Assertions
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)

	// Verify Casbin state directly
	actualRoles := enforcer.GetRolesForUserInDomain(userSub, domain)
	assert.ElementsMatch(t, []string{"EDITOR", "VIEWER"}, actualRoles, "Casbin roles should match assigned roles")
}

func TestAuthService_AssignRolesToUser_RoleNotFound(t *testing.T) {
	authSvc, mockRoleRepo, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()
	userID := "user1"
	roleIDs := []uint{99} // Non-existent role
	domain := "project:proj1"

	// Mock repository interaction
	mockRoleRepo.On("GetByID", ctx, uint(99)).Return(nil, errors.New("not found")).Once()

	// Call the service method
	err := authSvc.AssignRolesToUser(ctx, userID, roleIDs, domain)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "角色ID 99 不存在")
	mockRoleRepo.AssertExpectations(t)
}

func TestAuthService_IsUserInRole(t *testing.T) {
	authSvc, _, _, _, enforcer := setupAuthServiceTest(t)
	ctx := context.Background()
	userID := "user-abc"
	roleCode := "ADMIN"
	domain := "system"
	userSub := fmt.Sprintf("user:%s", userID)

	// Setup Casbin state for the test
	_, err := enforcer.AddRoleForUserInDomain(userSub, roleCode, domain)
	assert.NoError(t, err, "Failed to set up enforcer state for test")

	// Call the service method
	hasRole, err := authSvc.IsUserInRole(ctx, userID, roleCode, domain)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, hasRole)
}

func TestAuthService_IsUserInRole_NotInRole(t *testing.T) {
	authSvc, _, _, _, _ := setupAuthServiceTest(t) // Enforcer is initialized empty
	ctx := context.Background()
	userID := "user-abc"
	roleCode := "MEMBER"
	domain := "project:p123"

	// Call the service method (enforcer has no policy for this user/role)
	hasRole, err := authSvc.IsUserInRole(ctx, userID, roleCode, domain)

	// Assertions
	assert.NoError(t, err)
	assert.False(t, hasRole)
}

// --- Placeholder for future tests ---

// TODO: Add tests for RemoveRolesFromUser (similar logic to AssignRolesToUser)
// func TestAuthService_RemoveRolesFromUser(t *testing.T) { ... }

// TODO: Add tests for CanUserAccessResource
// func TestAuthService_CanUserAccessResource_DirectAllow(t *testing.T) { ... }
// func TestAuthService_CanUserAccessResource_RoleAllow(t *testing.T) { ... }
// func TestAuthService_CanUserAccessResource_Deny(t *testing.T) { ... }

// TODO: Add tests for other relevant AuthService methods (e.g., CreateRole, DeleteRole checking Casbin interaction)
// func TestAuthService_CreateRole(t *testing.T) { ... }

// TestAuthService_Basic runs a minimal test to verify we can instantiate AuthService with mocks
func TestAuthService_Basic(t *testing.T) {
	// Create mock repositories
	mockRoleRepo := mocks.NewRoleRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)
	mockCasbinRepo := mocks.NewCasbinRepository(t)

	// Create a simple in-memory Casbin enforcer
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	m, err := model.NewModelFromString(modelText)
	assert.NoError(t, err)

	adapter := fileadapter.NewAdapter("")
	enforcer, err := casbin.NewEnforcer(m, adapter)
	assert.NoError(t, err)

	// Create the service with our mocks
	authSvc := service.NewAuthService(
		enforcer,
		mockRoleRepo,
		mockUserRepo,
		mockCasbinRepo,
		nil, // pass nil for DB
	)

	// Set up a simple mock expectation
	ctx := context.Background()
	userID := "user123"

	// Expecting IsUserInRole to check if the user has the role
	// This will manually return false since we haven't added any roles
	hasRole, err := authSvc.IsUserInRole(ctx, userID, "ADMIN", "system")

	// Verify the result
	assert.NoError(t, err)
	assert.False(t, hasRole, "User should not have role")
}
