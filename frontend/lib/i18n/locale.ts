import { cookies } from "next/headers";
import { isLocale, type Locale } from "./messages";

const COOKIE_NAME = "locale";

export async function getServerLocale(): Promise<Locale> {
  const store = await cookies();
  const value = store.get(COOKIE_NAME)?.value;
  if (isLocale(value)) {
    return value;
  }
  return "en";
}

export const localeCookieName = COOKIE_NAME;
