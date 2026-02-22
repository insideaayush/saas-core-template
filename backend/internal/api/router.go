package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"saas-core-template/backend/internal/analytics"
	"saas-core-template/backend/internal/audit"
	"saas-core-template/backend/internal/auth"
	"saas-core-template/backend/internal/billing"
	"saas-core-template/backend/internal/files"
	"saas-core-template/backend/internal/orgs"
)

type Server struct {
	appName    string
	env        string
	version    string
	appBaseURL string
	db         *pgxpool.Pool
	redis      *redis.Client
	auth       *auth.Service
	billing    *billing.Service
	analytics  analytics.Client
	audit      audit.Recorder
	files      *files.Service
	orgs       *orgs.Service
}

type serverOptions struct {
	authService    *auth.Service
	billingService *billing.Service
	appBaseURL     string
	analytics      analytics.Client
	audit          audit.Recorder
	files          *files.Service
	orgs           *orgs.Service
}

func NewServer(appName string, env string, version string, db *pgxpool.Pool, redisClient *redis.Client, opts ...func(*serverOptions)) *Server {
	options := serverOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	return &Server{
		appName:    appName,
		env:        env,
		version:    version,
		appBaseURL: strings.TrimRight(defaultString(options.appBaseURL, "http://localhost:3000"), "/"),
		db:         db,
		redis:      redisClient,
		auth:       options.authService,
		billing:    options.billingService,
		analytics:  defaultAnalytics(options.analytics),
		audit:      defaultAudit(options.audit),
		files:      options.files,
		orgs:       options.orgs,
	}
}

func defaultAnalytics(client analytics.Client) analytics.Client {
	if client == nil {
		return analytics.NewNoop()
	}
	return client
}

func defaultAudit(recorder audit.Recorder) audit.Recorder {
	if recorder == nil {
		return audit.NewNoop()
	}
	return recorder
}

func WithAuthService(authService *auth.Service) func(*serverOptions) {
	return func(opts *serverOptions) {
		opts.authService = authService
	}
}

func WithBillingService(billingService *billing.Service) func(*serverOptions) {
	return func(opts *serverOptions) {
		opts.billingService = billingService
	}
}

func WithAppBaseURL(appBaseURL string) func(*serverOptions) {
	return func(opts *serverOptions) {
		opts.appBaseURL = strings.TrimSpace(appBaseURL)
	}
}

func WithAnalytics(client analytics.Client) func(*serverOptions) {
	return func(opts *serverOptions) {
		opts.analytics = client
	}
}

func WithAudit(recorder audit.Recorder) func(*serverOptions) {
	return func(opts *serverOptions) {
		opts.audit = recorder
	}
}

func WithFiles(service *files.Service) func(*serverOptions) {
	return func(opts *serverOptions) {
		opts.files = service
	}
}

func WithOrgs(service *orgs.Service) func(*serverOptions) {
	return func(opts *serverOptions) {
		opts.orgs = service
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.HandleFunc("GET /api/v1/meta", s.meta)
	mux.HandleFunc("GET /api/v1/auth/me", s.requireAuth(s.authMe))
	mux.HandleFunc("GET /api/v1/orgs", s.requireAuth(s.orgsList))
	mux.HandleFunc("POST /api/v1/orgs", s.requireAuth(s.orgsCreate))
	mux.HandleFunc("GET /api/v1/org/members", s.requireOrgRole(orgRoleAdmin, s.orgMembersList))
	mux.HandleFunc("POST /api/v1/org/invites", s.requireOrgRole(orgRoleAdmin, s.orgInvitesCreate))
	mux.HandleFunc("POST /api/v1/org/invites/accept", s.requireAuth(s.orgInvitesAccept))
	mux.HandleFunc("PATCH /api/v1/org/members/{userId}", s.requireOrgRole(orgRoleOwner, s.orgMembersUpdateRole))
	mux.HandleFunc("DELETE /api/v1/org/members/{userId}", s.requireOrgRole(orgRoleOwner, s.orgMembersRemove))
	mux.HandleFunc("POST /api/v1/billing/checkout-session", s.requireOrgRole(orgRoleAdmin, s.billingCheckoutSession))
	mux.HandleFunc("POST /api/v1/billing/portal-session", s.requireOrgRole(orgRoleAdmin, s.billingPortalSession))
	mux.HandleFunc("POST /api/v1/billing/webhook", s.billingWebhook)
	mux.HandleFunc("GET /api/v1/audit/events", s.requireOrgRole(orgRoleAdmin, s.auditEvents))
	mux.HandleFunc("POST /api/v1/files/upload-url", s.requireOrg(s.filesUploadURL))
	mux.HandleFunc("POST /api/v1/files/{id}/upload", s.requireOrg(s.filesDirectUpload))
	mux.HandleFunc("POST /api/v1/files/{id}/complete", s.requireOrg(s.filesComplete))
	mux.HandleFunc("GET /api/v1/files/{id}/download-url", s.requireOrg(s.filesDownloadURL))
	mux.HandleFunc("GET /api/v1/files/{id}/download", s.requireOrg(s.filesDownload))

	return withCommonMiddleware(mux)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not_ready",
			"reason": "database_unreachable",
		})
		return
	}

	if err := s.redis.Ping(ctx).Err(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not_ready",
			"reason": "redis_unreachable",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) meta(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"app":     s.appName,
		"env":     s.env,
		"version": s.version,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func withCommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Organization-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type authContextKey string

const (
	authUserContextKey authContextKey = "auth_user"
	authOrgContextKey  authContextKey = "auth_org"
)

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.auth == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth_not_configured"})
			return
		}

		token, err := auth.ExtractBearerToken(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing_or_invalid_token"})
			return
		}

		user, err := s.auth.Authenticate(r.Context(), token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication_failed"})
			return
		}

		s.analytics.Track(r.Context(), analytics.Event{
			Name:       "auth_authenticated",
			DistinctID: user.ID,
			Properties: map[string]any{"provider": "clerk"},
		})

		ctx := context.WithValue(r.Context(), authUserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) requireOrg(next http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		user := authUserFromContext(r.Context())
		if user.ID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user_not_found"})
			return
		}

		requestedOrgID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))
		org, err := s.auth.ResolveOrganization(r.Context(), user.ID, requestedOrgID)
		if err != nil {
			if errors.Is(err, auth.ErrNoOrganization) {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_not_found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "organization_resolution_failed"})
			return
		}

		ctx := context.WithValue(r.Context(), authOrgContextKey, org)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) authMe(w http.ResponseWriter, r *http.Request) {
	user := authUserFromContext(r.Context())
	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		resolved, err := s.auth.ResolveOrganization(r.Context(), user.ID, "")
		if err == nil {
			org = resolved
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":         user,
		"organization": org,
	})
}

func (s *Server) billingCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if s.billing == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "billing_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	var req struct {
		PlanCode string `json:"planCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	priceID, err := s.billing.LookupPlanPriceID(r.Context(), req.PlanCode)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown_or_inactive_plan"})
		return
	}

	customerID, _ := s.billing.GetOrganizationCustomerID(r.Context(), org.ID)
	session, err := s.billing.CreateCheckoutSession(r.Context(), billing.CheckoutSessionInput{
		OrganizationID: org.ID,
		CustomerID:     customerID,
		PriceID:        priceID,
		SuccessURL:     s.defaultSuccessURL(),
		CancelURL:      s.defaultPricingURL(),
	})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed_to_create_checkout_session"})
		return
	}

	s.analytics.Track(r.Context(), analytics.Event{
		Name:       "billing_checkout_session_created",
		DistinctID: user.ID,
		Properties: map[string]any{
			"organization_id": org.ID,
			"plan_code":       req.PlanCode,
		},
	})
	_ = s.audit.Record(r.Context(), audit.Event{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Action:         "billing_checkout_session_created",
		Data:           map[string]any{"plan_code": req.PlanCode},
	})

	writeJSON(w, http.StatusOK, map[string]string{"url": session.URL})
}

func (s *Server) billingPortalSession(w http.ResponseWriter, r *http.Request) {
	if s.billing == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "billing_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	customerID, _ := s.billing.GetOrganizationCustomerID(r.Context(), org.ID)
	if customerID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "organization_has_no_customer"})
		return
	}

	session, err := s.billing.CreatePortalSession(r.Context(), billing.PortalSessionInput{
		CustomerID: customerID,
		ReturnURL:  s.defaultAppURL(),
	})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed_to_create_portal_session"})
		return
	}

	s.analytics.Track(r.Context(), analytics.Event{
		Name:       "billing_portal_session_created",
		DistinctID: user.ID,
		Properties: map[string]any{
			"organization_id": org.ID,
		},
	})
	_ = s.audit.Record(r.Context(), audit.Event{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Action:         "billing_portal_session_created",
		Data:           map[string]any{},
	})

	writeJSON(w, http.StatusOK, map[string]string{"url": session.URL})
}

func (s *Server) billingWebhook(w http.ResponseWriter, r *http.Request) {
	if s.billing == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "billing_not_configured"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_webhook_payload"})
		return
	}

	if err := s.billing.VerifyWebhookSignature(r.Header.Get("Stripe-Signature"), body); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_webhook_signature"})
		return
	}

	if err := s.billing.HandleWebhookEvent(r.Context(), body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_process_webhook"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

func authUserFromContext(ctx context.Context) auth.User {
	user, ok := ctx.Value(authUserContextKey).(auth.User)
	if !ok {
		return auth.User{}
	}
	return user
}

func authOrgFromContext(ctx context.Context) auth.Organization {
	org, ok := ctx.Value(authOrgContextKey).(auth.Organization)
	if !ok {
		return auth.Organization{}
	}
	return org
}

func (s *Server) defaultPricingURL() string {
	return s.appBaseURL + "/pricing"
}

func (s *Server) defaultSuccessURL() string {
	return s.appBaseURL + "/app?billing=success"
}

func (s *Server) defaultAppURL() string {
	return s.appBaseURL + "/app"
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) auditEvents(w http.ResponseWriter, r *http.Request) {
	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	reader, ok := s.audit.(audit.Reader)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "audit_not_configured"})
		return
	}

	events, err := reader.ListByOrganization(r.Context(), org.ID, 50)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_list_audit_events"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

func (s *Server) filesUploadURL(w http.ResponseWriter, r *http.Request) {
	if s.files == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "files_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	var req struct {
		Filename    string `json:"filename"`
		ContentType string `json:"contentType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	resp, err := s.files.CreateUploadURL(r.Context(), files.CreateInput{
		OrganizationID: org.ID,
		UploaderUserID: user.ID,
		Filename:       req.Filename,
		ContentType:    req.ContentType,
	}, requestBaseURL(r))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_create_upload_url"})
		return
	}

	s.analytics.Track(r.Context(), analytics.Event{
		Name:       "file_upload_url_created",
		DistinctID: user.ID,
		Properties: map[string]any{"organization_id": org.ID},
	})
	_ = s.audit.Record(r.Context(), audit.Event{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Action:         "file_upload_url_created",
		Data:           map[string]any{"filename": req.Filename},
	})

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) filesDirectUpload(w http.ResponseWriter, r *http.Request) {
	if s.files == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "files_not_configured"})
		return
	}

	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	fileID := strings.TrimSpace(r.PathValue("id"))
	if fileID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_file_id"})
		return
	}

	if err := r.ParseMultipartForm(25 * 1024 * 1024); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_multipart_form"})
		return
	}

	f, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_file"})
		return
	}
	defer f.Close()

	if err := s.files.HandleDirectUpload(r.Context(), org.ID, fileID, f, header); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_upload_file"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "uploaded"})
}

func (s *Server) filesDownload(w http.ResponseWriter, r *http.Request) {
	if s.files == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "files_not_configured"})
		return
	}

	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	fileID := strings.TrimSpace(r.PathValue("id"))
	if fileID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_file_id"})
		return
	}

	url, err := s.files.GetDownloadURL(r.Context(), org.ID, fileID, requestBaseURL(r))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_get_download_url"})
		return
	}

	// Direct download for local disk provider.
	if strings.HasPrefix(url, requestBaseURL(r)+"/api/v1/files/") {
		if err := s.files.ServeDirectDownload(w, r, org.ID, fileID); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_download_file"})
			return
		}
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Server) filesDownloadURL(w http.ResponseWriter, r *http.Request) {
	if s.files == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "files_not_configured"})
		return
	}

	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	fileID := strings.TrimSpace(r.PathValue("id"))
	if fileID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_file_id"})
		return
	}

	base := requestBaseURL(r)
	url, err := s.files.GetDownloadURL(r.Context(), org.ID, fileID, base)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_get_download_url"})
		return
	}

	downloadType := "presigned"
	if strings.HasPrefix(url, base+"/api/v1/files/") {
		downloadType = "direct"
	}

	writeJSON(w, http.StatusOK, map[string]any{"url": url, "downloadType": downloadType})
}

func (s *Server) filesComplete(w http.ResponseWriter, r *http.Request) {
	if s.files == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "files_not_configured"})
		return
	}

	user := authUserFromContext(r.Context())
	org := authOrgFromContext(r.Context())
	if org.ID == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "organization_required"})
		return
	}

	fileID := strings.TrimSpace(r.PathValue("id"))
	if fileID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_file_id"})
		return
	}

	var req struct {
		SizeBytes int64 `json:"sizeBytes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := s.files.MarkUploaded(r.Context(), org.ID, fileID, req.SizeBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_mark_uploaded"})
		return
	}

	s.analytics.Track(r.Context(), analytics.Event{
		Name:       "file_uploaded",
		DistinctID: user.ID,
		Properties: map[string]any{"organization_id": org.ID},
	})
	_ = s.audit.Record(r.Context(), audit.Event{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Action:         "file_uploaded",
		Data:           map[string]any{"file_id": fileID, "size_bytes": req.SizeBytes},
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "uploaded"})
}

func requestBaseURL(r *http.Request) string {
	proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		proto = "http"
	}

	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}

	return proto + "://" + host
}
