import { cookies } from "next/headers";
import { isLocale, type Locale } from "./messages";

const COOKIE_NAME = "locale";

export function getServerLocale(): Locale {
  const value = cookies().get(COOKIE_NAME)?.value;
  if (isLocale(value)) {
    return value;
  }
  return "en";
}

export const localeCookieName = COOKIE_NAME;

