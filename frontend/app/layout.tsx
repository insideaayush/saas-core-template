import type { Metadata } from "next";
import "./globals.css";
import { AppClerkProvider } from "./clerk-provider";
import { AppIntegrationsProvider } from "./integrations-provider";
import { getServerLocale } from "@/lib/i18n/locale";
import { LanguageSwitcher } from "./language-switcher";

export const metadata: Metadata = {
  title: "SaaS Core Template",
  description: "Startup-ready SaaS template with auth, multi-tenant, and billing foundations."
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = getServerLocale();

  return (
    <html lang={locale}>
      <body>
        <div style={{ display: "flex", justifyContent: "flex-end", padding: "0.75rem 1rem" }}>
          <LanguageSwitcher currentLocale={locale} />
        </div>
        <AppClerkProvider>
          <AppIntegrationsProvider>{children}</AppIntegrationsProvider>
        </AppClerkProvider>
      </body>
    </html>
  );
}
