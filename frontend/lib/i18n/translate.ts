import type { Messages } from "./messages";

export function t(messages: Messages, key: string): string {
  const parts = key.split(".");
  let current: any = messages;
  for (const part of parts) {
    current = current?.[part];
  }

  if (typeof current === "string") {
    return current;
  }

  return key;
}
