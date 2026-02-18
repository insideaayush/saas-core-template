import { SignUp } from "@clerk/nextjs";

export default function SignUpPage() {
  return (
    <main>
      <h1>Create account</h1>
      <p>Get started quickly and spin up your workspace in minutes.</p>
      <div className="card" style={{ marginTop: "1rem" }}>
        <SignUp forceRedirectUrl="/app" />
      </div>
    </main>
  );
}
