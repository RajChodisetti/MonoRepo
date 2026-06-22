package auth

import "testing"

func TestValidRole(t *testing.T) {
	t.Parallel()

	for _, role := range []string{RoleAdmin, RoleUser, RoleDeveloper} {
		if !ValidRole(role) {
			t.Fatalf("ValidRole(%q) = false, want true", role)
		}
	}

	if ValidRole("superuser") {
		t.Fatal("ValidRole(superuser) = true, want false")
	}
}
