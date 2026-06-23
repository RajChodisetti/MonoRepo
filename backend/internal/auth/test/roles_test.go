package auth_test

import (
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

func TestValidRole(t *testing.T) {
	t.Parallel()

	for _, role := range []string{auth.RoleInternalAdmin, auth.RoleRestaurantOwner, auth.RoleDeveloper} {
		if !auth.ValidRole(role) {
			t.Fatalf("ValidRole(%q) = false, want true", role)
		}
	}

	if auth.ValidRole("superuser") {
		t.Fatal("ValidRole(superuser) = true, want false")
	}
}

func TestSignupAllowedRole(t *testing.T) {
	t.Parallel()

	if !auth.SignupAllowedRole(auth.RoleRestaurantOwner, "production") {
		t.Fatal("restaurant_owner signup should be allowed in production")
	}
	if auth.SignupAllowedRole(auth.RoleInternalAdmin, "local") {
		t.Fatal("internal_admin signup should not be allowed")
	}
	if !auth.SignupAllowedRole(auth.RoleDeveloper, "local") {
		t.Fatal("developer signup should be allowed in local")
	}
	if auth.SignupAllowedRole(auth.RoleDeveloper, "production") {
		t.Fatal("developer signup should not be allowed in production")
	}
}

func TestRoleHelpers(t *testing.T) {
	t.Parallel()

	if !auth.IsInternalAdmin(auth.RoleInternalAdmin) {
		t.Fatal("IsInternalAdmin(internal_admin) = false, want true")
	}
	if !auth.IsRestaurantOwner(auth.RoleRestaurantOwner) {
		t.Fatal("IsRestaurantOwner(restaurant_owner) = false, want true")
	}
}
