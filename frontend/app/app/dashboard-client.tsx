"use client";

import { UserButton, useAuth } from "@clerk/nextjs";
import { useEffect, useMemo, useState } from "react";
import {
  completeFileUpload,
  createBillingPortalSession,
  createFileUploadURL,
  fetchAuditEvents,
  fetchViewer,
  getFileDownloadURL,
  type AuditEventRecord,
  type ViewerResponse
} from "@/lib/api";
import { createAnalyticsClient } from "@/lib/integrations/analytics";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type LoadState = "idle" | "loading" | "error";

export function DashboardClient() {
  const { isLoaded, getToken, orgId, userId } = useAuth();
  const [viewer, setViewer] = useState<ViewerResponse | null>(null);
  const [state, setState] = useState<LoadState>("idle");
  const [portalLoading, setPortalLoading] = useState(false);
  const [auditEvents, setAuditEvents] = useState<AuditEventRecord[]>([]);
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [lastUploadedFileId, setLastUploadedFileId] = useState<string | null>(null);
  const analytics = useMemo(
    () => createAnalyticsClient((process.env.NEXT_PUBLIC_ANALYTICS_PROVIDER ?? "console") as "console" | "posthog" | "none"),
    []
  );

  const hasClerk = useMemo(() => Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY), []);

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

      const data = await fetchViewer(token, orgId);
      if (!cancelled) {
        setViewer(data);
        setState(data ? "idle" : "error");
      }
    }

    void loadViewer();
    return () => {
      cancelled = true;
    };
  }, [getToken, hasClerk, isLoaded, orgId, userId]);

  useEffect(() => {
    let cancelled = false;

    async function loadAudit() {
      if (!hasClerk) return;
      if (!isLoaded || !userId) return;
      const token = await getToken();
      if (!token) return;

      const data = await fetchAuditEvents(token, orgId);
      if (!cancelled) {
        setAuditEvents(data?.events ?? []);
      }
    }

    void loadAudit();
    return () => {
      cancelled = true;
    };
  }, [getToken, hasClerk, isLoaded, orgId, userId]);

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
    const session = await createBillingPortalSession({ token, organizationId: orgId });
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
      organizationId: orgId,
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
          ...(orgId ? { "X-Organization-ID": orgId } : {})
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
          organizationId: orgId,
          fileId: created.fileId,
          sizeBytes: uploadFile.size
        });
      }
    }

    setUploading(false);
    if (ok) {
      setLastUploadedFileId(created.fileId);
      analytics.track("file_uploaded", { fileId: created.fileId });
      const updated = await fetchAuditEvents(token, orgId);
      setAuditEvents(updated?.events ?? []);
    }
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
            <ul className="list-disc space-y-1 pl-5">
              <li>User: {viewer.user.primaryEmail || viewer.user.id}</li>
              <li>Organization: {viewer.organization.name || viewer.organization.id}</li>
              <li>Role: {viewer.organization.role}</li>
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
                  const info = await getFileDownloadURL({ token, organizationId: orgId, fileId: lastUploadedFileId });
                  if (!info) return;

                  if (info.downloadType === "presigned") {
                    window.open(info.url, "_blank", "noopener,noreferrer");
                    return;
                  }

                  const response = await fetch(info.url, {
                    method: "GET",
                    headers: {
                      Authorization: `Bearer ${token}`,
                      ...(orgId ? { "X-Organization-ID": orgId } : {})
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
