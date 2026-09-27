import { NextResponse, type NextRequest } from "next/server";
import { financialApiFetch, FinancialApiError } from "@/lib/api/financial-api";

type RouteContext = { params: Promise<{ path: string[] }> };

async function forward(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  const targetPath = `/v1/${path.join("/")}${request.nextUrl.search}`;
  const method = request.method;
  const body = method === "GET" || method === "HEAD" ? undefined : await request.text();
  const contentType = request.headers.get("content-type");

  try {
    const payload = await financialApiFetch<unknown>(targetPath, {
      method,
      body: body || undefined,
      headers: contentType ? { "Content-Type": contentType } : undefined,
    });
    return NextResponse.json(payload);
  } catch (error) {
    const status = error instanceof FinancialApiError ? error.status : 503;
    const message = error instanceof Error ? error.message : "Financial API unavailable";
    return NextResponse.json({ error: { code: "api_request_failed", message } }, { status });
  }
}

export const GET = forward;
export const POST = forward;
export const PATCH = forward;
export const PUT = forward;
