import type { Messages } from "./messages";

type DotPrefix<T extends string> = T extends "" ? "" : `.${T}`;
type DotNestedKeys<T> = T extends object
  ? {
      [K in Extract<keyof T, string>]: T[K] extends string
        ? `${K}`
        : T[K] extends readonly any[]
          ? `${K}`
          : `${K}${DotPrefix<DotNestedKeys<T[K]>>}`;
    }[Extract<keyof T, string>]
  : "";

export type MessageKey = DotNestedKeys<Messages>;

export function t(messages: Messages, key: MessageKey): string {
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

