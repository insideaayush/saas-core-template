"use client";

import { UserButton, useAuth } from "@clerk/nextjs";
import { useEffect, useMemo, useState } from "react";
import { createBillingPortalSession, fetchViewer, type ViewerResponse } from "@/lib/api";
import { createAnalyticsClient } from "@/lib/integrations/analytics";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type LoadState = "idle" | "loading" | "error";

export function DashboardClient() {
  const { isLoaded, getToken, orgId, userId } = useAuth();
  const [viewer, setViewer] = useState<ViewerResponse | null>(null);
  const [state, setState] = useState<LoadState>("idle");
  const [portalLoading, setPortalLoading] = useState(false);
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
    </div>
  );
}
