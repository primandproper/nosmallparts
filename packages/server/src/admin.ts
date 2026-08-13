/**
 * Admin status is derived from the ADMIN_EMAILS env var at request time
 * (comma-separated, case-insensitive). There is no DB column.
 */
export function isAdminEmail(email: string | null | undefined): boolean {
  if (!email) return false;
  const admins = (process.env.ADMIN_EMAILS || '')
    .split(',')
    .map((s) => s.trim().toLowerCase())
    .filter(Boolean);
  return admins.includes(email.trim().toLowerCase());
}
