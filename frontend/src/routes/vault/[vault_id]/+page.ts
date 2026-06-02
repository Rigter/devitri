export const prerender = false;

export function load({ params }: { params: { vault_id: string } }) {
  return { vault_id: params.vault_id };
}
