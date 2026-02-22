"use client";

import { useAuth } from "@clerk/nextjs";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { acceptOrganizationInvite } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

type State = "idle" | "accepting" | "accepted" | "error";

export function InviteClient() {
  const { isLoaded, getToken, userId } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const inviteToken = useMemo(() => searchParams.get("token") ?? "", [searchParams]);
  const [state, setState] = useState<State>("idle");

  useEffect(() => {
    let cancelled = false;

    async function accept() {
      if (!isLoaded || !userId) return;
      if (!inviteToken) {
        setState("error");
        return;
      }

      setState("accepting");
      const token = await getToken();
      if (!token) {
        setState("error");
        return;
      }

      const accepted = await acceptOrganizationInvite({ token, inviteToken });
      if (!accepted) {
        if (!cancelled) setState("error");
        return;
      }

      try {
        window.localStorage.setItem("activeOrganizationId", accepted.organization.id);
      } catch {
        // ignore
      }

      if (!cancelled) {
        setState("accepted");
        router.replace("/app");
      }
    }

    void accept();
    return () => {
      cancelled = true;
    };
  }, [getToken, inviteToken, isLoaded, router, userId]);

  return (
    <div className="mx-auto max-w-xl space-y-6 py-10">
      <Card>
        <CardHeader>
          <CardTitle>Accept Invite</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-muted-foreground">
          {state === "idle" && <p>Preparing to accept invite…</p>}
          {state === "accepting" && <p>Accepting your invite…</p>}
          {state === "accepted" && <p>Invite accepted. Redirecting…</p>}
          {state === "error" && (
            <div className="space-y-3">
              <p>Could not accept this invite. It may be invalid, already used, or intended for a different email.</p>
              <Button type="button" onClick={() => router.replace("/app")}>
                Back to app
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

