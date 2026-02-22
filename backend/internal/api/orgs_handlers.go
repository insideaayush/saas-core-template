package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"saas-core-template/backend/internal/analytics"
	"saas-core-template/backend/internal/orgs"
)

func (s *Server) orgsList(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "orgs_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	if user.ID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user_not_found"})
		return
	}

	items, err := s.orgs.ListForUser(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_list_orgs"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"organizations": items})
}

func (s *Server) orgsCreate(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "orgs_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	if user.ID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user_not_found"})
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	created, err := s.orgs.CreateTeamOrganization(r.Context(), user.ID, orgs.CreateOrgInput{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_create_org"})
		return
	}

	s.analytics.Track(r.Context(), analytics.Event{
		Name:       "organization_created",
		DistinctID: user.ID,
		Properties: map[string]any{"organization_id": created.ID, "kind": "team"},
	})

	writeJSON(w, http.StatusOK, map[string]any{"organization": created})
}

func (s *Server) orgMembersList(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "orgs_not_configured"})
		return
	}

	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	members, err := s.orgs.ListMembers(r.Context(), org.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_list_members"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"members": members})
}

func (s *Server) orgInvitesCreate(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "orgs_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	org := authOrgFromContext(r.Context())
	if user.ID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user_not_found"})
		return
	}
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	invite, err := s.orgs.CreateInvite(r.Context(), orgs.CreateInviteInput{
		OrganizationID:  org.ID,
		InvitedByUserID: user.ID,
		Email:           req.Email,
		Role:            req.Role,
	})
	if err != nil {
		if errors.Is(err, orgs.ErrInviteAlreadyExists) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "invite_already_exists"})
			return
		}
		if errors.Is(err, orgs.ErrInvalidOrganization) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invites_not_allowed_for_personal_workspace"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_create_invite"})
		return
	}

	acceptURL := strings.TrimRight(s.appBaseURL, "/") + "/app/invite?token=" + invite.Token
	_ = s.orgs.EnqueueInviteEmail(r.Context(), invite, acceptURL)
	writeJSON(w, http.StatusOK, map[string]any{
		"invite":    invite,
		"acceptUrl": acceptURL,
	})
}

func (s *Server) orgInvitesAccept(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "orgs_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	if user.ID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user_not_found"})
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	org, err := s.orgs.AcceptInvite(r.Context(), orgs.AcceptInviteInput{
		Token:  strings.TrimSpace(req.Token),
		UserID: user.ID,
		Email:  user.PrimaryEmail,
	})
	if err != nil {
		switch {
		case errors.Is(err, orgs.ErrInviteAlreadyUsed):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invite_already_used"})
		case errors.Is(err, orgs.ErrInviteEmailMismatch):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "invite_email_mismatch"})
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_accept_invite"})
		}
		return
	}

	s.analytics.Track(r.Context(), analytics.Event{
		Name:       "organization_invite_accepted",
		DistinctID: user.ID,
		Properties: map[string]any{"organization_id": org.ID},
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"organization": org,
		"acceptedAt":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) orgMembersUpdateRole(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "orgs_not_configured"})
		return
	}

	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	targetUserID := strings.TrimSpace(r.PathValue("userId"))
	if targetUserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_user_id"})
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	if err := s.orgs.UpdateMemberRole(r.Context(), orgs.UpdateMemberRoleInput{
		OrganizationID: org.ID,
		UserID:         targetUserID,
		Role:           req.Role,
	}); err != nil {
		if errors.Is(err, orgs.ErrInvalidOrganization) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role_changes_not_allowed_for_personal_workspace"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_update_role"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) orgMembersRemove(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "orgs_not_configured"})
		return
	}

	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	targetUserID := strings.TrimSpace(r.PathValue("userId"))
	if targetUserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_user_id"})
		return
	}

	if err := s.orgs.RemoveMember(r.Context(), orgs.RemoveMemberInput{
		OrganizationID: org.ID,
		UserID:         targetUserID,
	}); err != nil {
		if errors.Is(err, orgs.ErrInvalidOrganization) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "member_changes_not_allowed_for_personal_workspace"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_remove_member"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
