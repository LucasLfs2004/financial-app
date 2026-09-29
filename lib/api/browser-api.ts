import { API_NOTICE_EVENT, type ApiNotice } from "@/lib/api/notice-event";

function notify(detail: ApiNotice) {
  if (typeof window !== "undefined") window.dispatchEvent(new CustomEvent(API_NOTICE_EVENT, { detail }));
}

export class BrowserApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "BrowserApiError";
  }
}

export async function browserApiFetch<T>(path: string, init?: RequestInit & { successMessage?: string | false }): Promise<T> {
  const { successMessage, ...requestInit } = init ?? {};
  let response: Response;
  try {
    response = await fetch(`/api/financial${path}`, {
      ...requestInit,
      cache: "no-store",
      headers: {
        ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
        ...init?.headers,
      },
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") throw error;
    const message = "Sem conexão com o serviço. Tente novamente.";
    notify({ type: "error", message });
    throw new BrowserApiError(0, message);
  }
  const payload = (await response.json().catch(() => null)) as (T & { error?: { message?: string } }) | null;

  if (!response.ok) {
    const message = payload?.error?.message ?? "Não foi possível concluir a operação.";
    notify({ type: "error", message });
    throw new BrowserApiError(response.status, message);
  }

  if (init?.method && !["GET", "HEAD"].includes(init.method.toUpperCase()) && successMessage !== false) {
    notify({ type: "success", message: successMessage || "Operação concluída com sucesso." });
  }
  return payload as T;
}
