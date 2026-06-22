package auth

const (
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleDeveloper = "developer"
)

func ValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleUser, RoleDeveloper:
		return true
	default:
		return false
	}
}
