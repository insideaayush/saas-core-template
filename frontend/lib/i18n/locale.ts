import { cookies } from "next/headers";
import { isLocale, type Locale } from "./messages";

import { localeCookieName } from "./locale-cookie";

export async function getServerLocale(): Promise<Locale> {
  const store = await cookies();
  const value = store.get(localeCookieName)?.value;
  if (isLocale(value)) {
    return value;
  }
  return "en";
}
