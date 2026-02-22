"use client";

import { useAuth } from "@clerk/nextjs";
import { PropsWithChildren, useEffect, useMemo, useRef } from "react";
import { createAnalyticsClient } from "@/lib/integrations/analytics";
import { maybeLoadPostHog } from "@/lib/integrations/posthog-loader";
import { maybeLoadCrisp } from "@/lib/integrations/crisp-loader";
import { createSupportClient } from "@/lib/integrations/support";
import { createErrorReportingClient } from "@/lib/integrations/error-reporting";

type IntegrationClients = {
  analytics: ReturnType<typeof createAnalyticsClient>;
  support: ReturnType<typeof createSupportClient>;
  errorReporting: ReturnType<typeof createErrorReportingClient>;
};

function useIntegrationsBase(): IntegrationClients {
  const analyticsProvider = (process.env.NEXT_PUBLIC_ANALYTICS_PROVIDER ?? "console") as "console" | "posthog" | "none";
  const supportProvider = (process.env.NEXT_PUBLIC_SUPPORT_PROVIDER ?? "none") as "crisp" | "none";
  const errorReportingProvider = (process.env.NEXT_PUBLIC_ERROR_REPORTING_PROVIDER ?? "console") as "console" | "sentry" | "none";

  const clients = useRef<IntegrationClients | null>(null);

  const analytics = useMemo(() => createAnalyticsClient(analyticsProvider), [analyticsProvider]);
  const support = useMemo(() => createSupportClient(supportProvider), [supportProvider]);
  const errorReporting = useMemo(() => createErrorReportingClient(errorReportingProvider), [errorReportingProvider]);

  if (
    !clients.current ||
    clients.current.analytics !== analytics ||
    clients.current.support !== support ||
    clients.current.errorReporting !== errorReporting
  ) {
    clients.current = {
      analytics,
      support,
      errorReporting
    };
  }

  useEffect(() => {
    maybeLoadPostHog({
      provider: analyticsProvider,
      apiKey: process.env.NEXT_PUBLIC_POSTHOG_KEY,
      host: process.env.NEXT_PUBLIC_POSTHOG_HOST
    });

    maybeLoadCrisp({
      provider: supportProvider,
      websiteId: process.env.NEXT_PUBLIC_CRISP_WEBSITE_ID
    });

    clients.current?.errorReporting.init({
      dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
      environment: process.env.NEXT_PUBLIC_SENTRY_ENVIRONMENT,
      release: process.env.NEXT_PUBLIC_APP_VERSION
    });
  }, [analyticsProvider, errorReportingProvider, supportProvider]);

  return clients.current;
}

function IntegrationsWithClerk({ children }: PropsWithChildren) {
  const { isLoaded, userId, orgId } = useAuth();
  const integrations = useIntegrationsBase();

  useEffect(() => {
    integrations.analytics.page();
  }, [integrations.analytics]);

  useEffect(() => {
    if (!isLoaded || !userId) {
      return;
    }

    integrations.analytics.identify(userId);
    integrations.analytics.group("organization", orgId ?? "none");
    integrations.support.identify({ userId, organizationId: orgId ?? undefined });
    integrations.errorReporting.setUser({ id: userId });
  }, [integrations, isLoaded, orgId, userId]);

  return <>{children}</>;
}

export function AppIntegrationsProvider({ children }: PropsWithChildren) {
  const hasClerk = Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY);
  const integrations = useIntegrationsBase();

  useEffect(() => {
    integrations.analytics.page();
  }, [integrations.analytics]);

  if (!hasClerk) {
    return <>{children}</>;
  }

  return <IntegrationsWithClerk>{children}</IntegrationsWithClerk>;
}
