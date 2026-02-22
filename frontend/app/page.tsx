import Link from "next/link";
import { fetchMeta } from "@/lib/api";
import { getServerLocale } from "@/lib/i18n/locale";
import { getMessages } from "@/lib/i18n/messages";
import { t } from "@/lib/i18n/translate";

export default async function HomePage() {
  const meta = await fetchMeta();
  const hasClerk = Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY);
  const messages = getMessages(getServerLocale());

  return (
    <main>
      <h1>{t(messages, "home.title")}</h1>
      <p>{t(messages, "home.subtitle")}</p>

      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>{t(messages, "home.whatYouGetTitle")}</h2>
        <ul>
          {messages.home.whatYouGetBullets.map((bullet) => (
            <li key={bullet}>{bullet}</li>
          ))}
        </ul>
        <p>
          <Link href="/pricing">{t(messages, "home.whatYouGetCtaPrefix")} {t(messages, "home.whatYouGetCtaPricing")}</Link>{" "}
          {t(messages, "home.whatYouGetCtaOr")} <Link href="/app">{t(messages, "home.whatYouGetCtaDashboard")}</Link>.
        </p>
      </section>

      <section className="card" style={{ marginTop: "1.25rem" }}>
        <h2>{t(messages, "home.statusTitle")}</h2>
        {meta ? (
          <ul>
            <li>app: {meta.app}</li>
            <li>env: {meta.env}</li>
            <li>version: {meta.version}</li>
            <li>time: {new Date(meta.time).toLocaleString()}</li>
          </ul>
        ) : (
          <p>{t(messages, "home.backendUnreachable")}</p>
        )}
      </section>

      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>{t(messages, "home.getStartedTitle")}</h2>
        {hasClerk ? (
          <p>
            {t(messages, "home.getStartedWithClerk")}{" "}
            <Link href="/sign-up">sign up</Link> / <Link href="/sign-in">sign in</Link>
          </p>
        ) : (
          <p>
            {t(messages, "home.getStartedWithoutClerk")} <Link href="/app">/app</Link>
          </p>
        )}
      </section>
    </main>
  );
}
