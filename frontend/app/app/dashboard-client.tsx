"use client";

import { UserButton, useAuth } from "@clerk/nextjs";
import { useEffect, useMemo, useState } from "react";
import {
  completeFileUpload,
  createOrganization,
  createOrganizationInvite,
  createBillingPortalSession,
  createFileUploadURL,
  fetchAuditEvents,
  fetchOrganizationMembers,
  fetchOrganizations,
  fetchViewer,
  getFileDownloadURL,
  type AuditEventRecord,
  type OrganizationSummary,
  type OrganizationMembersResponse,
  type ViewerResponse
} from "@/lib/api";
import { createAnalyticsClient } from "@/lib/integrations/analytics";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type LoadState = "idle" | "loading" | "error";

function roleRank(role: string): number {
  switch ((role ?? "").toLowerCase()) {
    case "owner":
      return 3;
    case "admin":
      return 2;
    case "member":
      return 1;
    default:
      return 0;
  }
}

export function DashboardClient() {
  const { isLoaded, getToken, userId } = useAuth();
  const [viewer, setViewer] = useState<ViewerResponse | null>(null);
  const [organizations, setOrganizations] = useState<OrganizationSummary[]>([]);
  const [activeOrgId, setActiveOrgId] = useState<string | null>(null);
  const [state, setState] = useState<LoadState>("idle");
  const [portalLoading, setPortalLoading] = useState(false);
  const [auditEvents, setAuditEvents] = useState<AuditEventRecord[]>([]);
  const [members, setMembers] = useState<OrganizationMembersResponse["members"]>([]);
  const [membersState, setMembersState] = useState<LoadState>("idle");
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [lastUploadedFileId, setLastUploadedFileId] = useState<string | null>(null);
  const [newOrgName, setNewOrgName] = useState("");
  const [creatingOrg, setCreatingOrg] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<"member" | "admin">("member");
  const [inviteLink, setInviteLink] = useState<string | null>(null);
  const [inviteLoading, setInviteLoading] = useState(false);
  const analytics = useMemo(
    () => createAnalyticsClient((process.env.NEXT_PUBLIC_ANALYTICS_PROVIDER ?? "console") as "console" | "posthog" | "none"),
    []
  );

  const hasClerk = useMemo(() => Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY), []);

  useEffect(() => {
    try {
      const stored = window.localStorage.getItem("activeOrganizationId");
      if (stored) {
        setActiveOrgId(stored);
      }
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function loadOrganizations() {
      if (!hasClerk) return;
      if (!isLoaded || !userId) return;
      const token = await getToken();
      if (!token) return;

      const data = await fetchOrganizations(token);
      if (cancelled) return;
      setOrganizations(data?.organizations ?? []);
    }

    void loadOrganizations();
    return () => {
      cancelled = true;
    };
  }, [getToken, hasClerk, isLoaded, userId]);

  useEffect(() => {
    if (!organizations.length || !activeOrgId) return;
    if (organizations.some((o) => o.id === activeOrgId)) return;

    setActiveOrgId(organizations[0]?.id ?? null);
  }, [activeOrgId, organizations]);

  useEffect(() => {
    let cancelled = false;

    async function loadViewer() {
      if (!hasClerk) {
        setState("idle");
        return;
      }
      if (!isLoaded || !userId) {
        return;
      }

      setState("loading");
      const token = await getToken();
      if (!token) {
        if (!cancelled) {
          setState("error");
        }
        return;
      }

      const data = await fetchViewer(token, activeOrgId);
      if (!cancelled) {
        setViewer(data);
        setState(data ? "idle" : "error");
        if (data?.organization?.id) {
          setActiveOrgId(data.organization.id);
          try {
            window.localStorage.setItem("activeOrganizationId", data.organization.id);
          } catch {
            // ignore
          }
        }
      }
    }

    void loadViewer();
    return () => {
      cancelled = true;
    };
  }, [activeOrgId, getToken, hasClerk, isLoaded, userId]);

  useEffect(() => {
    let cancelled = false;

    async function loadAudit() {
      if (!hasClerk) return;
      if (!isLoaded || !userId) return;
      const token = await getToken();
      if (!token) return;

      const data = await fetchAuditEvents(token, activeOrgId);
      if (!cancelled) {
        setAuditEvents(data?.events ?? []);
      }
    }

    void loadAudit();
    return () => {
      cancelled = true;
    };
  }, [activeOrgId, getToken, hasClerk, isLoaded, userId]);

  useEffect(() => {
    let cancelled = false;

    async function loadMembers() {
      if (!hasClerk) return;
      if (!isLoaded || !userId) return;
      if (!activeOrgId) return;

      setMembersState("loading");
      const token = await getToken();
      if (!token) {
        if (!cancelled) setMembersState("error");
        return;
      }

      const data = await fetchOrganizationMembers(token, activeOrgId);
      if (cancelled) return;

      if (!data) {
        setMembers([]);
        setMembersState("idle");
        return;
      }

      setMembers(data.members);
      setMembersState("idle");
    }

    void loadMembers();
    return () => {
      cancelled = true;
    };
  }, [activeOrgId, getToken, hasClerk, isLoaded, userId]);

  const openBillingPortal = async () => {
    if (!hasClerk) {
      return;
    }
    analytics.track("billing_portal_open_clicked");
    setPortalLoading(true);
    const token = await getToken();
    if (!token) {
      setPortalLoading(false);
      return;
    }
    const session = await createBillingPortalSession({ token, organizationId: activeOrgId });
    setPortalLoading(false);
    if (session?.url) {
      window.location.href = session.url;
    }
  };

  const startUpload = async () => {
    if (!hasClerk || !uploadFile) return;
    setUploading(true);
    setLastUploadedFileId(null);

    const token = await getToken();
    if (!token) {
      setUploading(false);
      return;
    }

    const created = await createFileUploadURL({
      token,
      organizationId: activeOrgId,
      filename: uploadFile.name,
      contentType: uploadFile.type || "application/octet-stream"
    });
    if (!created) {
      setUploading(false);
      return;
    }

    let ok = false;
    if (created.uploadType === "direct") {
      const form = new FormData();
      form.append("file", uploadFile, uploadFile.name);
      const response = await fetch(created.url, {
        method: created.method,
        headers: {
          Authorization: `Bearer ${token}`,
          ...(activeOrgId ? { "X-Organization-ID": activeOrgId } : {})
        },
        body: form
      });
      ok = response.ok;
    } else {
      const response = await fetch(created.url, {
        method: created.method,
        headers: created.headers,
        body: uploadFile
      });
      ok = response.ok;
      if (ok) {
        ok = await completeFileUpload({
          token,
          organizationId: activeOrgId,
          fileId: created.fileId,
          sizeBytes: uploadFile.size
        });
      }
    }

    setUploading(false);
    if (ok) {
      setLastUploadedFileId(created.fileId);
      analytics.track("file_uploaded", { fileId: created.fileId });
      const updated = await fetchAuditEvents(token, activeOrgId);
      setAuditEvents(updated?.events ?? []);
    }
  };

  const createTeamOrg = async () => {
    if (!hasClerk) return;
    if (!newOrgName.trim()) return;
    const token = await getToken();
    if (!token) return;

    setCreatingOrg(true);
    const created = await createOrganization({ token, name: newOrgName.trim() });
    setCreatingOrg(false);
    if (!created?.organization?.id) return;

    setNewOrgName("");
    setActiveOrgId(created.organization.id);
    try {
      window.localStorage.setItem("activeOrganizationId", created.organization.id);
    } catch {
      // ignore
    }

    const updated = await fetchOrganizations(token);
    setOrganizations(updated?.organizations ?? []);
  };

  const createInvite = async () => {
    if (!hasClerk) return;
    if (!activeOrgId) return;
    if (!inviteEmail.trim()) return;
    const token = await getToken();
    if (!token) return;

    setInviteLoading(true);
    setInviteLink(null);
    const resp = await createOrganizationInvite({
      token,
      organizationId: activeOrgId,
      email: inviteEmail.trim(),
      role: inviteRole
    });
    setInviteLoading(false);
    if (!resp?.acceptUrl) return;
    setInviteLink(resp.acceptUrl);
  };

  return (
    <div className="space-y-6">
      <section className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">App Dashboard</h1>
        <p className="text-muted-foreground">Protected workspace for authenticated organizations.</p>
      </section>

      {hasClerk && (
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <div className="space-y-1">
              <CardTitle>Session</CardTitle>
              <p className="text-sm text-muted-foreground">{userId ? "Signed in with Clerk" : "Not signed in"}</p>
            </div>
            <UserButton afterSignOutUrl="/" />
          </CardHeader>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Workspace</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          {!hasClerk && (
            <p>
              Clerk is not configured yet. Set <code>NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY</code> to enable auth and organization context.
            </p>
          )}
          {hasClerk && state === "loading" && <p>Loading your workspace context...</p>}
          {hasClerk && state === "error" && <p>Could not load workspace context from API. Ensure backend auth and migrations are configured.</p>}
          {hasClerk && viewer && (
            <div className="space-y-3">
              <ul className="list-disc space-y-1 pl-5">
                <li>User: {viewer.user.primaryEmail || viewer.user.id}</li>
                <li>
                  Active organization: {viewer.organization.name || viewer.organization.id}{" "}
                  <span className="text-xs">({viewer.organization.kind})</span>
                </li>
                <li>Role: {viewer.organization.role}</li>
              </ul>

              <div className="space-y-1">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Switch workspace</p>
                <select
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm text-foreground"
                  value={activeOrgId ?? ""}
                  onChange={(e) => {
                    const next = e.target.value || null;
                    setActiveOrgId(next);
                    setInviteLink(null);
                    try {
                      if (next) window.localStorage.setItem("activeOrganizationId", next);
                    } catch {
                      // ignore
                    }
                  }}
                >
                  {(organizations.length ? organizations : [viewer.organization]).map((org) => (
                    <option key={org.id} value={org.id}>
                      {org.name || org.slug || org.id} ({org.kind}, {org.role})
                    </option>
                  ))}
                </select>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Organizations</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
          <p>Create a team workspace and invite members.</p>
          <div className="flex flex-col gap-2 sm:flex-row">
            <input
              className="w-full rounded-md border bg-background px-3 py-2 text-sm text-foreground"
              placeholder="New team workspace name"
              value={newOrgName}
              onChange={(e) => setNewOrgName(e.target.value)}
            />
            <Button type="button" onClick={createTeamOrg} disabled={!hasClerk || creatingOrg || !newOrgName.trim()}>
              {creatingOrg ? "Creating..." : "Create team org"}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Invites</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
          <p>Invite a teammate (team orgs only, admin+).</p>
          <div className="flex flex-col gap-2 sm:flex-row">
            <input
              className="w-full rounded-md border bg-background px-3 py-2 text-sm text-foreground"
              placeholder="teammate@example.com"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
            />
            <select
              className="rounded-md border bg-background px-3 py-2 text-sm text-foreground"
              value={inviteRole}
              onChange={(e) => setInviteRole((e.target.value as "member" | "admin") ?? "member")}
            >
              <option value="member">member</option>
              <option value="admin">admin</option>
            </select>
            <Button
              type="button"
              onClick={createInvite}
              disabled={!hasClerk || inviteLoading || !inviteEmail.trim() || roleRank(viewer?.organization.role ?? "") < 2}
            >
              {inviteLoading ? "Creating..." : "Create invite"}
            </Button>
          </div>
          {inviteLink && (
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">Invite link</p>
              <code className="block break-all rounded-md bg-muted px-3 py-2 text-xs text-foreground">{inviteLink}</code>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          {roleRank(viewer?.organization.role ?? "") < 2 ? (
            <p>Members are visible to admins and owners.</p>
          ) : membersState === "loading" ? (
            <p>Loading members...</p>
          ) : members.length === 0 ? (
            <p>No members found.</p>
          ) : (
            <ul className="space-y-2">
              {members.map((m) => (
                <li key={m.userId}>
                  <span className="font-medium text-foreground">{m.primaryEmail || m.userId}</span>{" "}
                  <span className="text-xs">({m.role})</span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Billing</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
          <p>Manage your plan and invoices in the Stripe customer portal.</p>
          <div>
            <Button type="button" onClick={openBillingPortal} disabled={!hasClerk || portalLoading}>
              {portalLoading ? "Opening..." : "Open billing portal"}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Files</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
          <p>Tenant-scoped uploads with disk local storage or presigned S3/R2 URLs.</p>
          <input
            type="file"
            onChange={(e) => {
              setUploadFile(e.target.files?.[0] ?? null);
            }}
          />
          <div className="flex items-center gap-3">
            <Button type="button" onClick={startUpload} disabled={!hasClerk || uploading || !uploadFile}>
              {uploading ? "Uploading..." : "Upload"}
            </Button>
            {lastUploadedFileId && (
              <Button
                type="button"
                variant="outline"
                onClick={async () => {
                  const token = await getToken();
                  if (!token) return;
                  const info = await getFileDownloadURL({ token, organizationId: activeOrgId, fileId: lastUploadedFileId });
                  if (!info) return;

                  if (info.downloadType === "presigned") {
                    window.open(info.url, "_blank", "noopener,noreferrer");
                    return;
                  }

                  const response = await fetch(info.url, {
                    method: "GET",
                    headers: {
                      Authorization: `Bearer ${token}`,
                      ...(activeOrgId ? { "X-Organization-ID": activeOrgId } : {})
                    }
                  });
                  if (!response.ok) return;
                  const blob = await response.blob();
                  const href = URL.createObjectURL(blob);
                  const a = document.createElement("a");
                  a.href = href;
                  a.download = "download";
                  a.click();
                  URL.revokeObjectURL(href);
                }}
              >
                Download last upload
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Audit events</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          {auditEvents.length === 0 ? (
            <p>No recent events.</p>
          ) : (
            <ul className="space-y-2">
              {auditEvents.slice(0, 10).map((evt) => (
                <li key={evt.id}>
                  <span className="font-medium text-foreground">{evt.action}</span>{" "}
                  <span className="text-xs">({new Date(evt.createdAt).toLocaleString()})</span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
