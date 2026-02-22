"use client";

import { useRouter } from "next/navigation";
import { useMemo } from "react";
import { LOCALES, type Locale } from "@/lib/i18n/messages";
import { localeCookieName } from "@/lib/i18n/locale";

export function LanguageSwitcher({ currentLocale }: { currentLocale: Locale }) {
  const router = useRouter();

  const options = useMemo(
    () =>
      LOCALES.map((locale) => ({
        value: locale,
        label: locale.toUpperCase()
      })),
    []
  );

  return (
    <label style={{ display: "inline-flex", gap: "0.5rem", alignItems: "center" }}>
      <span style={{ fontSize: "0.9rem", opacity: 0.8 }}>Lang</span>
      <select
        value={currentLocale}
        onChange={(e) => {
          const next = e.target.value as Locale;
          document.cookie = `${localeCookieName}=${next}; Path=/; SameSite=Lax; Max-Age=31536000`;
          router.refresh();
        }}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </label>
  );
}

