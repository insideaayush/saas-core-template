import Link from "next/link";
import { PricingClient } from "./pricing-client";

export default function PricingPage() {
  return (
    <main>
      <h1>Pricing</h1>
      <p>Simple plans designed to get from zero to production without custom billing plumbing.</p>
      <PricingClient />
      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>Need help choosing?</h2>
        <p>
          Start with Pro and upgrade anytime. If you are already onboarded, head to the <Link href="/app">dashboard</Link>.
        </p>
      </section>
    </main>
  );
}
