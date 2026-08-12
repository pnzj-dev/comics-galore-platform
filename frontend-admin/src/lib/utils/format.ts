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

export function formatCurrency(cents: number | undefined | null): string {
	if (cents === undefined || cents === null || isNaN(cents)) return '—';
	const usd = cents / 100;
	if (usd >= 1000000) return `$${(usd / 1000000).toFixed(1)}M`;
	if (usd >= 1000) return `$${(usd / 1000).toFixed(1)}K`;
	return `$${usd.toFixed(0)}`;
}

export function formatNum(n: number | undefined | null): string {
	if (n === undefined || n === null || isNaN(n)) return '—';
	if (n >= 1000000) return `${(n / 1000000).toFixed(1)}M`;
	if (n >= 1000) return `${(n / 1000).toFixed(1)}K`;
	return String(n);
}

export function formatBytes(bytes: number | undefined | null): string {
	if (bytes === undefined || bytes === null || isNaN(bytes)) return '—';
	if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(1)} GB`;
	if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(1)} MB`;
	if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${bytes} B`;
}


