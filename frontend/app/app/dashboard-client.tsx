"use client";

import { UserButton, useAuth } from "@clerk/nextjs";
import { useEffect, useMemo, useState } from "react";
import { createBillingPortalSession, fetchViewer, type ViewerResponse } from "@/lib/api";
import { createAnalyticsClient } from "@/lib/integrations/analytics";

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
    <main>
      <h1>App Dashboard</h1>
      <p>Protected workspace for authenticated organizations.</p>

      {hasClerk && (
        <div className="card" style={{ marginTop: "1rem", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <div>
            <h2 style={{ marginTop: 0 }}>Session</h2>
            <p style={{ marginBottom: 0 }}>{userId ? "Signed in with Clerk" : "Not signed in"}</p>
          </div>
          <UserButton afterSignOutUrl="/" />
        </div>
      )}

      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>Workspace</h2>
        {!hasClerk && (
          <p>
            Clerk is not configured yet. Set <code>NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY</code> to enable auth and organization context.
          </p>
        )}
        {hasClerk && state === "loading" && <p>Loading your workspace context...</p>}
        {hasClerk && state === "error" && <p>Could not load workspace context from API. Ensure backend auth and migrations are configured.</p>}
        {hasClerk && viewer && (
          <ul>
            <li>User: {viewer.user.primaryEmail || viewer.user.id}</li>
            <li>Organization: {viewer.organization.name || viewer.organization.id}</li>
            <li>Role: {viewer.organization.role}</li>
          </ul>
        )}
      </section>

      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>Billing</h2>
        <p>Manage your plan and invoices in the Stripe customer portal.</p>
        <button type="button" onClick={openBillingPortal} disabled={!hasClerk || portalLoading}>
          {portalLoading ? "Opening..." : "Open billing portal"}
        </button>
      </section>
    </main>
  );
}
