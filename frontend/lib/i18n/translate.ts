import type { Messages } from "./messages";

export function t(messages: Messages, key: string): string {
  const parts = key.split(".");
  let current: unknown = messages;
  for (const part of parts) {
    if (!current || typeof current !== "object") {
      current = undefined;
      break;
    }

    const record = current as Record<string, unknown>;
    current = record[part];
  }

  if (typeof current === "string") {
    return current;
  }

  return key;
}
