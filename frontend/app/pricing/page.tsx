import Link from "next/link";
import { PricingClient } from "./pricing-client";
import { getServerLocale } from "@/lib/i18n/locale";
import { getMessages } from "@/lib/i18n/messages";
import { t } from "@/lib/i18n/translate";

export default function PricingPage() {
  const messages = getMessages(getServerLocale());

  return (
    <main>
      <h1>{t(messages, "pricing.title")}</h1>
      <p>{t(messages, "pricing.subtitle")}</p>
      <PricingClient />
      <section className="card" style={{ marginTop: "1rem" }}>
        <h2>{t(messages, "pricing.helpTitle")}</h2>
        <p>
          {t(messages, "pricing.helpBody")} <Link href="/app">dashboard</Link>.
        </p>
      </section>
    </main>
  );
}
