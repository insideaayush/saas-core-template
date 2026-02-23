export type ErrorReportingProvider = "console" | "sentry" | "none";

export type ErrorReportingClient = {
  init: (config: { dsn?: string; environment?: string; release?: string }) => void;
  captureException: (error: unknown, context?: Record<string, unknown>) => void;
  setUser: (user: { id: string } | null) => void;
};

export function createErrorReportingClient(provider: ErrorReportingProvider): ErrorReportingClient {
  if (provider === "none") {
    return {
      init: () => {},
      captureException: () => {},
      setUser: () => {}
    };
  }

  if (provider === "console") {
    return {
      init: () => {},
      captureException: (error, context) => {
        console.error("[error-reporting]", error, context);
      },
      setUser: (user) => {
        console.info("[error-reporting:user]", user);
      }
    };
  }

  // sentry (browser SDK via script loader)
  return {
    init: ({ dsn, environment, release }) => {
      maybeLoadSentryBrowser();

      const trimmed = dsn?.trim();
      if (!trimmed) {
        return;
      }

      window.Sentry?.init?.({
        dsn: trimmed,
        environment: environment?.trim() || undefined,
        release: release?.trim() || undefined
      });
    },
    captureException: (error, context) => {
      window.Sentry?.withScope?.((scope) => {
        if (context) {
          scope.setContext?.("extra", context);
        }
        window.Sentry?.captureException?.(error);
      });
    },
    setUser: (user) => {
      window.Sentry?.setUser?.(user);
    }
  };
}

function maybeLoadSentryBrowser() {
  if (window.Sentry?.init) {
    return;
  }

  if (document.querySelector('script[data-sentry-loader="true"]')) {
    return;
  }

  const script = document.createElement("script");
  script.async = true;
  script.defer = true;
  script.dataset.sentryLoader = "true";
  script.crossOrigin = "anonymous";
  // Keep version pinned for deterministic builds; update intentionally.
  script.src = "https://browser.sentry-cdn.com/7.120.0/bundle.tracing.min.js";

  document.head.appendChild(script);
}

declare global {
  interface Window {
    Sentry?: {
      init?: (config: { dsn: string; environment?: string; release?: string }) => void;
      withScope?: (fn: (scope: { setContext?: (name: string, data: Record<string, unknown>) => void }) => void) => void;
      captureException?: (error: unknown) => void;
      setUser?: (user: { id: string } | null) => void;
    };
  }
}
