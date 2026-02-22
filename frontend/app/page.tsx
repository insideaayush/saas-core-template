import Link from "next/link";
import { fetchMeta } from "@/lib/api";
import { getServerLocale } from "@/lib/i18n/locale";
import { getMessages } from "@/lib/i18n/messages";
import { t } from "@/lib/i18n/translate";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default async function HomePage() {
  const meta = await fetchMeta();
  const hasClerk = Boolean(process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY);
  const messages = getMessages(await getServerLocale());

  return (
    <div className="space-y-6">
      <section className="space-y-3">
        <h1 className="text-4xl font-semibold tracking-tight">{t(messages, "home.title")}</h1>
        <p className="max-w-2xl text-muted-foreground">{t(messages, "home.subtitle")}</p>
        <div className="flex flex-wrap gap-3 pt-2">
          <Button asChild>
            <Link href="/pricing">View pricing</Link>
          </Button>
          <Button variant="outline" asChild>
            <Link href="/app">Open app</Link>
          </Button>
        </div>
      </section>

      <Card>
        <CardHeader>
          <CardTitle>{t(messages, "home.whatYouGetTitle")}</CardTitle>
          <CardDescription>Template baseline features included out of the box.</CardDescription>
        </CardHeader>
        <CardContent>
          <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground">
            {messages.home.whatYouGetBullets.map((bullet) => (
              <li key={bullet}>{bullet}</li>
            ))}
          </ul>
          <div className="mt-4 text-sm">
            <Link className="underline underline-offset-4" href="/pricing">
              {t(messages, "home.whatYouGetCtaPrefix")} {t(messages, "home.whatYouGetCtaPricing")}
            </Link>{" "}
            <span className="text-muted-foreground">{t(messages, "home.whatYouGetCtaOr")}</span>{" "}
            <Link className="underline underline-offset-4" href="/app">
              {t(messages, "home.whatYouGetCtaDashboard")}
            </Link>
            .
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t(messages, "home.statusTitle")}</CardTitle>
          <CardDescription>Live backend connectivity check.</CardDescription>
        </CardHeader>
        <CardContent>
          {meta ? (
            <ul className="grid gap-2 text-sm text-muted-foreground">
              <li>app: {meta.app}</li>
              <li>env: {meta.env}</li>
              <li>version: {meta.version}</li>
              <li>time: {new Date(meta.time).toLocaleString()}</li>
            </ul>
          ) : (
            <p className="text-sm text-muted-foreground">{t(messages, "home.backendUnreachable")}</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t(messages, "home.getStartedTitle")}</CardTitle>
          <CardDescription>Auth and tenancy are optional until configured.</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          {hasClerk ? (
            <p>
              {t(messages, "home.getStartedWithClerk")}{" "}
              <Link className="underline underline-offset-4" href="/sign-up">
                sign up
              </Link>{" "}
              /{" "}
              <Link className="underline underline-offset-4" href="/sign-in">
                sign in
              </Link>
            </p>
          ) : (
            <p>
              {t(messages, "home.getStartedWithoutClerk")}{" "}
              <Link className="underline underline-offset-4" href="/app">
                /app
              </Link>
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
