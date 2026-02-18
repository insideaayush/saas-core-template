import { auth } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";
import type { ReactNode } from "react";

export default async function ProtectedAppLayout({
  children
}: {
  children: ReactNode;
}) {
  const hasPublishableKey = Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY);
  if (!hasPublishableKey) {
    return children;
  }

  const { userId } = await auth();
  if (!userId) {
    redirect("/sign-in");
  }

  return children;
}
