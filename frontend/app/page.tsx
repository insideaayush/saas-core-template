import Link from "next/link";
import { fetchMeta } from "@/lib/api";

export default async function HomePage() {
  const meta = await fetchMeta();
  const hasClerk = Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY);

  return (
    <main>
      <h1>SaaS Core Template</h1>
      <p>Launch a production-shaped SaaS baseline with auth, multi-tenant workspaces, and billing foundations.</p>

      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>What you get</h2>
        <ul>
          <li>Landing + pricing pages with clear upgrade paths</li>
          <li>Protected app area and organization-aware APIs</li>
          <li>Managed auth and billing integrations that stay migration-friendly</li>
        </ul>
        <p>
          <Link href="/pricing">See pricing</Link> or continue to the <Link href="/app">app dashboard</Link>.
        </p>
      </section>

      <section className="card" style={{ marginTop: "1.25rem" }}>
        <h2>Platform Status</h2>
        {meta ? (
          <ul>
            <li>app: {meta.app}</li>
            <li>env: {meta.env}</li>
            <li>version: {meta.version}</li>
            <li>time: {new Date(meta.time).toLocaleString()}</li>
          </ul>
        ) : (
          <p>Backend is unreachable. Start API on port 8080.</p>
        )}
      </section>

      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>Get started</h2>
        {hasClerk ? (
          <p>
            Use <Link href="/sign-up">sign up</Link> to create your account or <Link href="/sign-in">sign in</Link> if you already have one.
          </p>
        ) : (
          <p>
            Clerk is not configured yet. Add publishable and secret keys, then use <Link href="/app">/app</Link> as your protected dashboard.
          </p>
        )}
      </section>
    </main>
  );
}
