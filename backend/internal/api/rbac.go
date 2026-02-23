package api

import (
	"net/http"
	"strings"
)

type orgRole string

const (
	orgRoleMember orgRole = "member"
	orgRoleAdmin  orgRole = "admin"
	orgRoleOwner  orgRole = "owner"
)

func (r orgRole) rank() (int, bool) {
	switch strings.ToLower(strings.TrimSpace(string(r))) {
	case string(orgRoleMember):
		return 1, true
	case string(orgRoleAdmin):
		return 2, true
	case string(orgRoleOwner):
		return 3, true
	default:
		return 0, false
	}
}

func orgRoleAllows(actual string, required orgRole) bool {
	actualRank, ok := orgRole(actual).rank()
	if !ok {
		return false
	}

	requiredRank, ok := required.rank()
	if !ok {
		return false
	}

	return actualRank >= requiredRank
}

func (s *Server) requireOrgRole(required orgRole, next http.HandlerFunc) http.HandlerFunc {
	return s.requireOrg(func(w http.ResponseWriter, r *http.Request) {
		org := authOrgFromContext(r.Context())
		if org.ID == "" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
			return
		}

		if !orgRoleAllows(org.Role, required) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient_role"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
