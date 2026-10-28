import { API_NOTICE_EVENT, type ApiNotice } from "@/lib/api/notice-event";
import { apiErrorDetail, isRequestAborted } from "@/lib/api/errors";

export const API_SESSION_EXPIRED_EVENT = "financial-api-session-expired";

function notify(detail: ApiNotice) {
  if (typeof window !== "undefined") window.dispatchEvent(new CustomEvent(API_NOTICE_EVENT, { detail }));
}

export class BrowserApiError extends Error {
  constructor(public status: number, message: string, public code = "api_request_failed", public details?: unknown) {
    super(message);
    this.name = "BrowserApiError";
  }
}

export async function browserApiFetch<T>(path: string, init?: RequestInit & { successMessage?: string | false }): Promise<T> {
  const { successMessage, ...requestInit } = init ?? {};
  let response: Response;
  const headers = new Headers(requestInit.headers);
  if (requestInit.body instanceof FormData) headers.delete("Content-Type");
  else if (requestInit.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  try {
    response = await fetch(`/api/financial${path}`, {
      ...requestInit,
      cache: "no-store",
      headers,
    });
  } catch (error) {
    if (isRequestAborted(error)) throw error;
    const mutation = requestInit.method && !["GET", "HEAD"].includes(requestInit.method.toUpperCase());
    const message = mutation ? "A conexão foi interrompida. Atualize os dados para verificar se a alteração foi salva antes de enviar novamente." : "Sem conexão com o serviço. Tente novamente.";
    notify({ type: "error", message });
    throw new BrowserApiError(0, message, "network_error");
  }
  const payload = (await response.json().catch(() => null)) as (T & { error?: { message?: string } }) | null;

  if (!response.ok) {
    const detail = apiErrorDetail(payload, response.status);
    if (response.status === 401 && typeof window !== "undefined") window.dispatchEvent(new Event(API_SESSION_EXPIRED_EVENT));
    notify({ type: "error", message: response.status === 401 ? "Sua sessão expirou. Entre novamente." : detail.message });
    throw new BrowserApiError(response.status, detail.message, detail.code, detail.details);
  }

  if (init?.method && !["GET", "HEAD"].includes(init.method.toUpperCase()) && successMessage !== false) {
    notify({ type: "success", message: successMessage || "Operação concluída com sucesso." });
  }
  return payload as T;
}
