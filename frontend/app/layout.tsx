import type { Metadata } from "next";
import "./globals.css";
import { AppClerkProvider } from "./clerk-provider";
import { AppIntegrationsProvider } from "./integrations-provider";
import { getServerLocale } from "@/lib/i18n/locale";
import { LanguageSwitcher } from "./language-switcher";
import { Card } from "@/components/ui/card";

export const metadata: Metadata = {
  title: "SaaS Core Template",
  description: "Startup-ready SaaS template with auth, multi-tenant, and billing foundations."
};

export default async function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = await getServerLocale();

  return (
    <html lang={locale}>
      <body className="min-h-screen">
        <header className="sticky top-0 z-10 border-b bg-background/80 backdrop-blur">
          <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-3">
            <div className="text-sm font-semibold tracking-tight">SaaS Core Template</div>
            <div className="flex items-center gap-3">
              <LanguageSwitcher currentLocale={locale} />
            </div>
          </div>
        </header>
        <div className="mx-auto max-w-5xl px-6 py-8">
          <Card className="border-dashed bg-card/40">
            <div className="p-4 text-sm text-muted-foreground">
              This template uses shadcn/ui primitives. Replace this banner with your product nav.
            </div>
          </Card>
        </div>
        <AppClerkProvider>
          <AppIntegrationsProvider>
            <main className="mx-auto max-w-5xl px-6 pb-16">{children}</main>
          </AppIntegrationsProvider>
        </AppClerkProvider>
      </body>
    </html>
  );
}
