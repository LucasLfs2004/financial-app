export class BrowserApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "BrowserApiError";
  }
}

export async function browserApiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/financial${path}`, {
    ...init,
    headers: {
      ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...init?.headers,
    },
  });
  const payload = (await response.json().catch(() => null)) as T & { error?: { message?: string } };

  if (!response.ok) {
    throw new BrowserApiError(response.status, payload?.error?.message ?? "Não foi possível concluir a operação.");
  }

  return payload;
}
