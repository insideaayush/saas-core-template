import type { SupportProvider } from "./support";

export type CrispConfig = {
  provider: SupportProvider;
  websiteId?: string;
};

export function maybeLoadCrisp(config: CrispConfig): void {
  if (config.provider !== "crisp") {
    return;
  }

  const websiteId = config.websiteId?.trim();
  if (!websiteId) {
    return;
  }

  window.$crisp = window.$crisp ?? [];
  window.CRISP_WEBSITE_ID = websiteId;

  if (document.querySelector('script[data-crisp-loader="true"]')) {
    return;
  }

  const script = document.createElement("script");
  script.async = true;
  script.defer = true;
  script.dataset.crispLoader = "true";
  script.src = "https://client.crisp.chat/l.js";
  document.head.appendChild(script);
}

