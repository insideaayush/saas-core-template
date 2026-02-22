import type { Metadata } from "next";
import "./globals.css";
import { AppClerkProvider } from "./clerk-provider";
import { AppIntegrationsProvider } from "./integrations-provider";

export const metadata: Metadata = {
  title: "SaaS Core Template",
  description: "Startup-ready SaaS template with auth, multi-tenant, and billing foundations."
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AppClerkProvider>
          <AppIntegrationsProvider>{children}</AppIntegrationsProvider>
        </AppClerkProvider>
      </body>
    </html>
  );
}
