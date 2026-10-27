export type ApiErrorDetail = { code: string; message: string; details?: unknown };

export function apiErrorDetail(payload: unknown, status: number): ApiErrorDetail {
  const body = payload && typeof payload === "object" ? payload as Record<string, unknown> : {};
  const error = body.error && typeof body.error === "object" ? body.error as Record<string, unknown> : body;
  return {
    code: typeof error.code === "string" ? error.code : "api_request_failed",
    message: typeof body.error === "string" ? body.error : typeof error.message === "string" ? error.message : `Não foi possível concluir a operação (HTTP ${status}).`,
    ...(error.details !== undefined ? { details: error.details } : {}),
  };
}

export function isRequestAborted(error: unknown) {
  return error instanceof Error && error.name === "AbortError";
}
