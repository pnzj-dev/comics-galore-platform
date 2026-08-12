export function formatDate(d: string | undefined | null): string {
	if (!d) return '—';
	const date = new Date(d);
	if (isNaN(date.getTime())) return '—';
	const pad = (n: number) => String(n).padStart(2, '0');
	return `${pad(date.getDate())}/${pad(date.getMonth() + 1)}/${date.getFullYear()} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function formatUSD(cents: number | undefined | null): string {
	if (cents === undefined || cents === null || isNaN(cents)) return '—';
	return `$${(cents / 100).toFixed(2)}`;
}

