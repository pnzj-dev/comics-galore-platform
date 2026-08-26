export type DateVariant = 'short' | 'long' | 'datetime' | 'admin';

export function formatDate(
	d: string | undefined | null,
	variant: DateVariant = 'admin',
	fallback = '—'
): string {
	if (!d) return fallback;
	const date = new Date(d);
	if (isNaN(date.getTime())) return fallback;
	switch (variant) {
		case 'long':
			return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
		case 'datetime':
			return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
		case 'short':
			return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
		case 'admin':
		default: {
			const pad = (n: number) => String(n).padStart(2, '0');
			return `${pad(date.getDate())}/${pad(date.getMonth() + 1)}/${date.getFullYear()} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
		}
	}
}

export function formatNumber(n: number | undefined | null, fallback = '—'): string {
	if (n === undefined || n === null || isNaN(n)) return fallback;
	if (n >= 1000000) return `${(n / 1000000).toFixed(1).replace(/\.0$/, '')}M`;
	if (n >= 1000) return `${(n / 1000).toFixed(1).replace(/\.0$/, '')}k`;
	return String(n);
}

export function formatBytes(bytes: number | undefined | null, fallback = '—'): string {
	if (bytes === undefined || bytes === null || isNaN(bytes)) return fallback;
	if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(1)} GB`;
	if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(1)} MB`;
	if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${bytes} B`;
}

export function formatUSD(cents: number | undefined | null): string {
	if (cents === undefined || cents === null || isNaN(cents)) return '—';
	return `$${(cents / 100).toFixed(2)}`;
}

export function formatCurrency(cents: number | undefined | null): string {
	if (cents === undefined || cents === null || isNaN(cents)) return '—';
	const usd = cents / 100;
	if (usd >= 1000000) return `$${(usd / 1000000).toFixed(1)}M`;
	if (usd >= 1000) return `$${(usd / 1000).toFixed(1)}k`;
	return `$${usd.toFixed(0)}`;
}

export function statusColor(status: string): string {
	switch (status) {
		case 'published': return 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400';
		case 'pending_review': return 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400';
		case 'rejected': return 'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400';
		default: return 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400';
	}
}
