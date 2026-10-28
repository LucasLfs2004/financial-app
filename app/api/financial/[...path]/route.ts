import { NextResponse, type NextRequest } from "next/server";
import { financialApiResponse, FinancialApiError } from "@/lib/api/financial-api";

type RouteContext = { params: Promise<{ path: string[] }> };

async function forward(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  const targetPath = `/v1/${path.map(encodeURIComponent).join("/")}${request.nextUrl.search}`;
  const method = request.method;
  const body = method === "GET" || method === "HEAD" ? undefined : await request.arrayBuffer();
  const contentType = request.headers.get("content-type");

  try {
    const response = await financialApiResponse(targetPath, {
      method,
      body: body?.byteLength ? body : undefined,
      headers: contentType ? { "Content-Type": contentType } : undefined,
      signal: path.includes("imports") ? AbortSignal.timeout(30000) : undefined,
    });
    return new Response(response.body, {
      status: response.status,
      headers: { "Content-Type": response.headers.get("content-type") ?? "application/json", "Cache-Control": "no-store" },
    });
  } catch (error) {
    const status = error instanceof FinancialApiError ? error.status : 503;
    const message = error instanceof FinancialApiError ? error.message : "Serviço financeiro indisponível. Atualize os dados antes de repetir uma alteração.";
    const code = error instanceof FinancialApiError ? error.code : error instanceof Error && error.name === "TimeoutError" ? "api_timeout" : "api_unavailable";
    return NextResponse.json({ error: { code, message } }, { status });
  }
}

export const GET = forward;
export const POST = forward;
export const PATCH = forward;
export const PUT = forward;
