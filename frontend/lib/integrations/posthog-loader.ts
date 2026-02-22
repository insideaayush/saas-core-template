import type { AnalyticsProvider } from "./analytics";

export type PostHogConfig = {
  provider: AnalyticsProvider;
  apiKey?: string;
  host?: string;
};

export function maybeLoadPostHog(config: PostHogConfig): void {
  if (config.provider !== "posthog") {
    return;
  }

  const key = config.apiKey?.trim();
  if (!key) {
    return;
  }

  const host = (config.host?.trim() || "https://app.posthog.com").replace(/\/+$/, "");

  if (window.posthog?.init) {
    window.posthog.init(key, { api_host: host, capture_pageview: false });
    return;
  }

  if (document.querySelector('script[data-posthog-loader="true"]')) {
    return;
  }

  // Minimal loader: injects the official library and initializes it once loaded.
  // Avoids adding an npm dependency so local E2E works without installs.
  window.posthog = window.posthog ?? {};

  const script = document.createElement("script");
  script.async = true;
  script.defer = true;
  script.dataset.posthogLoader = "true";
  script.src = `${host}/static/array.js`;
  script.onload = () => {
    window.posthog?.init?.(key, { api_host: host, capture_pageview: false });
  };

  document.head.appendChild(script);
}

