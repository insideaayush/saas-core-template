"use client";

import { useRouter } from "next/navigation";
import { useMemo } from "react";
import { LOCALES, type Locale } from "@/lib/i18n/messages";
import { localeCookieName } from "@/lib/i18n/locale-cookie";
import { Button } from "@/components/ui/button";

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
    <div className="flex items-center gap-2">
      <span className="text-xs text-muted-foreground">Lang</span>
      <div className="flex rounded-md border bg-background p-1">
        {options.map((opt) => {
          const active = opt.value === currentLocale;
          return (
            <Button
              key={opt.value}
              type="button"
              size="sm"
              variant={active ? "secondary" : "ghost"}
              className="h-8 px-2"
              onClick={() => {
                const next = opt.value as Locale;
                document.cookie = `${localeCookieName}=${next}; Path=/; SameSite=Lax; Max-Age=31536000`;
                router.refresh();
              }}
            >
              {opt.label}
            </Button>
          );
        })}
      </div>
    </div>
  );
}
