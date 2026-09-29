export function formatDate(date: string): string {
  const match = /^(\d{4})-(\d{2})-(\d{2})(?:$|T)/.exec(date);
  return match ? `${match[3]}/${match[2]}/${match[1]}` : date;
}
