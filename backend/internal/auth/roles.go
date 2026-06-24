package auth

const (
	RoleInternalAdmin   = "internal_admin"
	RoleRestaurantOwner = "restaurant_owner"
	RoleDeveloper       = "developer"
)

func ValidRole(role string) bool {
	switch role {
	case RoleInternalAdmin, RoleRestaurantOwner, RoleDeveloper:
		return true
	default:
		return false
	}
}

// SignupAllowedRole controls which roles can be self-assigned via public signup.
// Internal admin accounts must be created through seed-admin or other internal flows.
func SignupAllowedRole(role, appEnv string) bool {
	switch role {
	case RoleRestaurantOwner:
		return true
	case RoleDeveloper:
		return appEnv == "local" || appEnv == "test"
	default:
		return false
	}
}

func IsInternalAdmin(role string) bool {
	return role == RoleInternalAdmin
}

func IsRestaurantOwner(role string) bool {
	return role == RoleRestaurantOwner
}
