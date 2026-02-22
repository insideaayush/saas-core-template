export const LOCALES = ["en", "es"] as const;
export type Locale = (typeof LOCALES)[number];

export type Messages = {
  home: {
    title: string;
    subtitle: string;
    whatYouGetTitle: string;
    whatYouGetBullets: [string, string, string];
    whatYouGetCtaPrefix: string;
    whatYouGetCtaPricing: string;
    whatYouGetCtaOr: string;
    whatYouGetCtaDashboard: string;
    statusTitle: string;
    backendUnreachable: string;
    getStartedTitle: string;
    getStartedWithClerk: string;
    getStartedWithoutClerk: string;
  };
  pricing: {
    title: string;
    subtitle: string;
    helpTitle: string;
    helpBody: string;
  };
  common: {
    language: string;
  };
};

const EN: Messages = {
  home: {
    title: "SaaS Core Template",
    subtitle: "Launch a production-shaped SaaS baseline with auth, multi-tenant workspaces, and billing foundations.",
    whatYouGetTitle: "What you get",
    whatYouGetBullets: [
      "Landing + pricing pages with clear upgrade paths",
      "Protected app area and organization-aware APIs",
      "Managed auth and billing integrations that stay migration-friendly"
    ],
    whatYouGetCtaPrefix: "See",
    whatYouGetCtaPricing: "pricing",
    whatYouGetCtaOr: "or continue to the",
    whatYouGetCtaDashboard: "app dashboard",
    statusTitle: "Platform Status",
    backendUnreachable: "Backend is unreachable. Start API on port 8080.",
    getStartedTitle: "Get started",
    getStartedWithClerk: "Use sign up to create your account or sign in if you already have one.",
    getStartedWithoutClerk: "Clerk is not configured yet. Add keys, then use /app as your protected dashboard."
  },
  pricing: {
    title: "Pricing",
    subtitle: "Simple plans designed to get from zero to production without custom billing plumbing.",
    helpTitle: "Need help choosing?",
    helpBody: "Start with Pro and upgrade anytime. If you are already onboarded, head to the dashboard."
  },
  common: {
    language: "Language"
  }
};

const ES: Messages = {
  home: {
    title: "Plantilla SaaS Core",
    subtitle: "Lanza una base SaaS lista para producción con autenticación, multi-tenant y facturación.",
    whatYouGetTitle: "Qué incluye",
    whatYouGetBullets: [
      "Landing + precios con rutas claras de upgrade",
      "Área protegida y APIs con contexto de organización",
      "Integraciones gestionadas de auth y billing con migración fácil"
    ],
    whatYouGetCtaPrefix: "Ver",
    whatYouGetCtaPricing: "precios",
    whatYouGetCtaOr: "o continuar al",
    whatYouGetCtaDashboard: "panel de la app",
    statusTitle: "Estado de la plataforma",
    backendUnreachable: "El backend no responde. Inicia el API en el puerto 8080.",
    getStartedTitle: "Empezar",
    getStartedWithClerk: "Usa registro para crear tu cuenta o iniciar sesión si ya tienes una.",
    getStartedWithoutClerk: "Clerk no está configurado. Agrega las claves y usa /app como panel protegido."
  },
  pricing: {
    title: "Precios",
    subtitle: "Planes simples para pasar de cero a producción sin construir facturación desde cero.",
    helpTitle: "¿Necesitas ayuda para elegir?",
    helpBody: "Empieza con Pro y actualiza cuando quieras. Si ya estás listo, ve al panel."
  },
  common: {
    language: "Idioma"
  }
};

export function isLocale(value: string | undefined | null): value is Locale {
  return (LOCALES as readonly string[]).includes(value ?? "");
}

export function getMessages(locale: Locale): Messages {
  return locale === "es" ? ES : EN;
}

