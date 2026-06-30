import { redirect } from "next/navigation";

export default async function RestaurantAliasPage({
  params,
}: {
  params: Promise<{ index: string }>;
}) {
  const { index } = await params;
  redirect(`/?id=${index}`);
}
