import { SignIn } from "@clerk/nextjs";

export default function SignInPage() {
  return (
    <main>
      <h1>Sign in</h1>
      <p>Access your workspace with magic links, social login, or your configured methods.</p>
      <div className="card" style={{ marginTop: "1rem" }}>
        <SignIn forceRedirectUrl="/app" />
      </div>
    </main>
  );
}
