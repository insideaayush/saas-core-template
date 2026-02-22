package api

import "testing"

func TestOrgRoleAllows(t *testing.T) {
	t.Run("rejects unknown actual role", func(t *testing.T) {
		if orgRoleAllows("superadmin", orgRoleMember) {
			t.Fatalf("expected false")
		}
	})

	t.Run("rejects unknown required role", func(t *testing.T) {
		if orgRoleAllows("owner", orgRole("weird")) {
			t.Fatalf("expected false")
		}
	})

	t.Run("member meets member", func(t *testing.T) {
		if !orgRoleAllows("member", orgRoleMember) {
			t.Fatalf("expected true")
		}
	})

	t.Run("member does not meet admin", func(t *testing.T) {
		if orgRoleAllows("member", orgRoleAdmin) {
			t.Fatalf("expected false")
		}
	})

	t.Run("admin meets member", func(t *testing.T) {
		if !orgRoleAllows("admin", orgRoleMember) {
			t.Fatalf("expected true")
		}
	})

	t.Run("owner meets admin", func(t *testing.T) {
		if !orgRoleAllows("owner", orgRoleAdmin) {
			t.Fatalf("expected true")
		}
	})
}
