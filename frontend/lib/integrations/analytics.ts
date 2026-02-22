export type AnalyticsProvider = "console" | "posthog" | "none";

export type AnalyticsClient = {
  identify: (userId: string, props?: Record<string, unknown>) => void;
  group: (groupType: string, groupKey: string, props?: Record<string, unknown>) => void;
  track: (event: string, props?: Record<string, unknown>) => void;
  page: (path?: string) => void;
};

function log(method: string, event: string, props?: Record<string, unknown>) {
  const payload = props ? JSON.stringify(props) : "";
  // eslint-disable-next-line no-console
  console.info(`[analytics:${method}] ${event}`, payload);
}

export function createAnalyticsClient(provider: AnalyticsProvider): AnalyticsClient {
  if (provider === "none") {
    return {
      identify: () => {},
      group: () => {},
      track: () => {},
      page: () => {}
    };
  }

  if (provider === "console") {
    return {
      identify: (userId, props) => log("identify", userId, props),
      group: (groupType, groupKey, props) => log("group", `${groupType}:${groupKey}`, props),
      track: (event, props) => log("track", event, props),
      page: (path) => log("page", path ?? window.location.pathname)
    };
  }

  // posthog
  return {
    identify: (userId, props) => {
      window.posthog?.identify?.(userId, props);
    },
    group: (groupType, groupKey, props) => {
      window.posthog?.group?.(groupType, groupKey, props);
    },
    track: (event, props) => {
      window.posthog?.capture?.(event, props);
    },
    page: (path) => {
      const url = path ?? window.location.pathname;
      window.posthog?.capture?.("$pageview", { $current_url: url });
    }
  };
}

declare global {
  interface Window {
    posthog?: {
      init?: (key: string, options: { api_host: string; capture_pageview?: boolean }) => void;
      identify?: (distinctId: string, props?: Record<string, unknown>) => void;
      group?: (groupType: string, groupKey: string, props?: Record<string, unknown>) => void;
      capture?: (event: string, props?: Record<string, unknown>) => void;
    };
  }
}

