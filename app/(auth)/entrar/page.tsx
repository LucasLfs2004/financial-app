import { redirect } from "next/navigation";

export default async function SignInPage({
  searchParams,
}: {
  searchParams: Promise<{ erro?: string; next?: string }>;
}) {
  const params = await searchParams;
  const query = new URLSearchParams();
  if (params.erro) query.set("erro", params.erro);
  if (params.next) query.set("next", params.next);
  redirect(`/sign-in${query.size ? `?${query.toString()}` : ""}`);
}
