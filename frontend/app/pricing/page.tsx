import Link from "next/link";
import { PricingClient } from "./pricing-client";
import { getServerLocale } from "@/lib/i18n/locale";
import { getMessages } from "@/lib/i18n/messages";
import { t } from "@/lib/i18n/translate";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default async function PricingPage() {
  const messages = getMessages(await getServerLocale());

  return (
    <div className="space-y-6">
      <section className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">{t(messages, "pricing.title")}</h1>
        <p className="text-muted-foreground">{t(messages, "pricing.subtitle")}</p>
      </section>
      <PricingClient />
      <Card>
        <CardHeader>
          <CardTitle>{t(messages, "pricing.helpTitle")}</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          {t(messages, "pricing.helpBody")}{" "}
          <Link className="underline underline-offset-4" href="/app">
            dashboard
          </Link>
          .
        </CardContent>
      </Card>
    </div>
  );
}
