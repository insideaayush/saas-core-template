import type { Metadata } from "next";
import "./globals.css";
import { AppClerkProvider } from "./clerk-provider";

export const metadata: Metadata = {
  title: "Novame WebOS",
  description: "Personal operating system shell for life and work."
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AppClerkProvider>{children}</AppClerkProvider>
      </body>
    </html>
  );
}
