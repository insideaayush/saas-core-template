"use client";

import { useAuth } from "@clerk/nextjs";
import { useState } from "react";
import { createCheckoutSession } from "@/lib/api";
import { createAnalyticsClient } from "@/lib/integrations/analytics";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const PLANS = [
  {
    code: "pro",
    name: "Pro",
    price: "$29",
    description: "For solo builders shipping quickly.",
    features: ["Single workspace", "Core collaboration tools", "Email support"]
  },
  {
    code: "team",
    name: "Team",
    price: "$99",
    description: "For growing teams with shared ownership.",
    features: ["Multiple members", "Role-based access", "Priority support"]
  }
] as const;

export function PricingClient() {
  const { getToken } = useAuth();
  const [loadingPlan, setLoadingPlan] = useState<string | null>(null);
  const hasClerk = Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY);
  const analytics = createAnalyticsClient((process.env.NEXT_PUBLIC_ANALYTICS_PROVIDER ?? "console") as "console" | "posthog" | "none");

  const startCheckout = async (planCode: string) => {
    if (!hasClerk) {
      return;
    }

    analytics.track("pricing_choose_plan_clicked", { planCode });
    setLoadingPlan(planCode);
    const token = await getToken();
    if (!token) {
      setLoadingPlan(null);
      window.location.href = "/sign-in";
      return;
    }

    const session = await createCheckoutSession({
      token,
      planCode,
      organizationId: (() => {
        try {
          return window.localStorage.getItem("activeOrganizationId");
        } catch {
          return null;
        }
      })()
    });

    setLoadingPlan(null);
    if (session?.url) {
      window.location.href = session.url;
    }
  };

  return (
    <section className="grid gap-4 md:grid-cols-2">
      {PLANS.map((plan) => (
        <Card key={plan.code}>
          <CardHeader>
            <CardTitle>{plan.name}</CardTitle>
            <div className="text-2xl font-semibold">
              {plan.price}
              <span className="text-sm font-normal text-muted-foreground">/mo</span>
            </div>
            <p className="text-sm text-muted-foreground">{plan.description}</p>
          </CardHeader>
          <CardContent>
            <ul className="list-disc space-y-1 pl-5 text-sm text-muted-foreground">
            {plan.features.map((feature) => (
              <li key={feature}>{feature}</li>
            ))}
            </ul>
            <div className="pt-4">
              <Button type="button" disabled={!hasClerk || loadingPlan === plan.code} onClick={() => void startCheckout(plan.code)}>
                {loadingPlan === plan.code ? "Redirecting..." : `Choose ${plan.name}`}
              </Button>
            </div>
          </CardContent>
        </Card>
      ))}
      {!hasClerk && (
        <Card className="md:col-span-2">
          <CardContent className="pt-6 text-sm text-muted-foreground">
            Pricing checkout requires Clerk session auth. Configure <code>NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY</code> and backend keys to enable checkout.
          </CardContent>
        </Card>
      )}
    </section>
  );
}
