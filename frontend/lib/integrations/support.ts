export type SupportProvider = "crisp" | "none";

export type SupportClient = {
  identify: (params: { userId?: string; email?: string; organizationId?: string }) => void;
  open: () => void;
};

export function createSupportClient(provider: SupportProvider): SupportClient {
  if (provider === "none") {
    return { identify: () => {}, open: () => {} };
  }

  // crisp
  return {
    identify: ({ userId, email, organizationId }) => {
      const crisp = window.$crisp;
      if (!crisp) {
        return;
      }

      if (email) {
        crisp.push(["set", "user:email", [email]]);
      }
      if (userId) {
        crisp.push(["set", "session:data", [[["user_id", userId]]]]);
      }
      if (organizationId) {
        crisp.push(["set", "session:data", [[["organization_id", organizationId]]]]);
      }
    },
    open: () => {
      window.$crisp?.push?.(["do", "chat:open"]);
    }
  };
}

declare global {
  interface Window {
    $crisp?: Array<unknown>;
    CRISP_WEBSITE_ID?: string;
  }
}
