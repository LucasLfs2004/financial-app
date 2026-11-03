import { FinancialApiError } from "./financial-api";

// Only known read failures become unavailable UI. Redirects/auth and unexpected
// programming errors must retain their original behavior.
export function unavailableRead(error: unknown): null {
  if (error instanceof FinancialApiError && error.status !== 401) return null;
  if (error instanceof TypeError && /fetch failed|failed to fetch|network/i.test(error.message)) return null;
  if (error instanceof DOMException && ["AbortError", "TimeoutError", "NetworkError"].includes(error.name)) return null;
  throw error;
}
