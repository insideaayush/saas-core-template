export type AppMeta = {
  app: string;
  env: string;
  version: string;
  time: string;
};

const FALLBACK_API_URL = "http://localhost:8080";
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? FALLBACK_API_URL;

export type ViewerResponse = {
  user: {
    id: string;
    primaryEmail: string;
  };
  organization: {
    id: string;
    name: string;
    slug: string;
    role: string;
  };
};

export type AuditEventRecord = {
  id: string;
  organizationId: string;
  userId: string;
  action: string;
  data: Record<string, unknown>;
  createdAt: string;
};

export type AuditEventsResponse = {
  events: AuditEventRecord[];
};

export type FileUploadURLResponse = {
  fileId: string;
  method: string;
  url: string;
  headers: Record<string, string>;
  uploadType: "direct" | "presigned";
};

export type FileDownloadURLResponse = {
  url: string;
  downloadType: "direct" | "presigned";
};

export async function fetchMeta(): Promise<AppMeta | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/meta`, {
      cache: "no-store"
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as AppMeta;
  } catch {
    return null;
  }
}

export async function fetchViewer(token: string, organizationId?: string | null): Promise<ViewerResponse | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/me`, {
      method: "GET",
      headers: buildAuthHeaders(token, organizationId)
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as ViewerResponse;
  } catch {
    return null;
  }
}

export async function fetchAuditEvents(token: string, organizationId?: string | null): Promise<AuditEventsResponse | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/audit/events`, {
      method: "GET",
      headers: buildAuthHeaders(token, organizationId)
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as AuditEventsResponse;
  } catch {
    return null;
  }
}

export async function createFileUploadURL(params: {
  token: string;
  organizationId?: string | null;
  filename: string;
  contentType: string;
}): Promise<FileUploadURLResponse | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/files/upload-url`, {
      method: "POST",
      headers: {
        ...buildAuthHeaders(params.token, params.organizationId),
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ filename: params.filename, contentType: params.contentType })
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as FileUploadURLResponse;
  } catch {
    return null;
  }
}

export async function completeFileUpload(params: {
  token: string;
  organizationId?: string | null;
  fileId: string;
  sizeBytes?: number;
}): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/files/${params.fileId}/complete`, {
      method: "POST",
      headers: {
        ...buildAuthHeaders(params.token, params.organizationId),
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ sizeBytes: params.sizeBytes ?? 0 })
    });

    return response.ok;
  } catch {
    return false;
  }
}

export async function getFileDownloadURL(params: {
  token: string;
  organizationId?: string | null;
  fileId: string;
}): Promise<FileDownloadURLResponse | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/files/${params.fileId}/download-url`, {
      method: "GET",
      headers: buildAuthHeaders(params.token, params.organizationId)
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as FileDownloadURLResponse;
  } catch {
    return null;
  }
}

export async function createCheckoutSession(params: {
  token: string;
  planCode: string;
  organizationId?: string | null;
}): Promise<{ url: string } | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/billing/checkout-session`, {
      method: "POST",
      headers: {
        ...buildAuthHeaders(params.token, params.organizationId),
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        planCode: params.planCode
      })
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as { url: string };
  } catch {
    return null;
  }
}

export async function createBillingPortalSession(params: {
  token: string;
  organizationId?: string | null;
}): Promise<{ url: string } | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/billing/portal-session`, {
      method: "POST",
      headers: buildAuthHeaders(params.token, params.organizationId)
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as { url: string };
  } catch {
    return null;
  }
}

function buildAuthHeaders(token: string, organizationId?: string | null): HeadersInit {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${token}`
  };

  if (organizationId) {
    headers["X-Organization-ID"] = organizationId;
  }

  return headers;
}
