"use client";

import { useAuth } from "@clerk/nextjs";
import { useState } from "react";
import { createCheckoutSession } from "@/lib/api";

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
  const { getToken, orgId } = useAuth();
  const [loadingPlan, setLoadingPlan] = useState<string | null>(null);
  const hasClerk = Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY);

  const startCheckout = async (planCode: string) => {
    if (!hasClerk) {
      return;
    }

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
      organizationId: orgId
    });

    setLoadingPlan(null);
    if (session?.url) {
      window.location.href = session.url;
    }
  };

  return (
    <section style={{ marginTop: "1rem", display: "grid", gap: "1rem", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))" }}>
      {PLANS.map((plan) => (
        <article key={plan.code} className="card">
          <h2>{plan.name}</h2>
          <p style={{ fontSize: "1.5rem", margin: "0.5rem 0" }}>{plan.price}/mo</p>
          <p>{plan.description}</p>
          <ul>
            {plan.features.map((feature) => (
              <li key={feature}>{feature}</li>
            ))}
          </ul>
          <button type="button" disabled={!hasClerk || loadingPlan === plan.code} onClick={() => void startCheckout(plan.code)}>
            {loadingPlan === plan.code ? "Redirecting..." : `Choose ${plan.name}`}
          </button>
        </article>
      ))}
      {!hasClerk && (
        <div className="card" style={{ gridColumn: "1 / -1" }}>
          <p>
            Pricing checkout requires Clerk session auth. Configure <code>NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY</code> and backend keys to enable checkout.
          </p>
        </div>
      )}
    </section>
  );
}
