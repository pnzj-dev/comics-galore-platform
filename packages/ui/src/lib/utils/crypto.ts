// Crypto currency helpers: base-symbol normalization and stablecoin detection.
// Token icons live in the CoinIcon component (svelte-cryptocurrency-icons).

const NETWORK_SUFFIXES = [
	'arbitrum',
	'avalanche',
	'optimism',
	'starknet',
	'polygon',
	'mantle',
	'gnosis',
	'moonbeam',
	'moonriver',
	'harmony',
	'klaytn',
	'cronos',
	'fantom',
	'scroll',
	'zksync',
	'linea',
	'aptos',
	'aurora',
	'tezos',
	'trc20',
	'erc20',
	'bep20',
	'avaxc',
	'bep2',
	'avax',
	'base',
	'celo',
	'near',
	'algo',
	'xdai',
	'evmos',
	'kava',
	'tron',
	'omni',
	'sol',
	'sei',
	'one',
	'bsc',
	'ton',
	'eos'
].sort((a, b) => b.length - a.length);

const BASE_SYMBOL_OVERRIDES: Record<string, string> = {
	usdttrc20: 'usdt',
	usdterc20: 'usdt',
	usdtbep20: 'usdt',
	usdtbsc: 'usdt',
	usdtpolygon: 'usdt',
	usdtsol: 'usdt',
	usdtavaxc: 'usdt',
	usdtarbitrum: 'usdt',
	usdtoptimism: 'usdt',
	usdtton: 'usdt',
	usdtnear: 'usdt',
	usdtbase: 'usdt',
	usdctrc20: 'usdc',
	usdcerc20: 'usdc',
	usdcbep20: 'usdc',
	usdcbsc: 'usdc',
	usdcpolygon: 'usdc',
	usdcsol: 'usdc',
	usdcavaxc: 'usdc',
	usdcarbitrum: 'usdc',
	usdcoptimism: 'usdc',
	usdcbase: 'usdc',
	usdccelo: 'usdc',
	usdcton: 'usdc',
	busderc20: 'busd',
	busdbep20: 'busd',
	busdbsc: 'busd',
	wbtc: 'wbtc',
	weth: 'weth'
};

const STABLECOINS = new Set([
	'usdt',
	'usdc',
	'tusd',
	'dai',
	'busd',
	'usdp',
	'gusd',
	'fdusd',
	'pyusd',
	'usdd',
	'usde',
	'usdn',
	'usdj',
	'usdh',
	'frax',
	'lusd',
	'mim',
	'ust',
	'susd',
	'cusd',
	'zusd',
	'eurt',
	'eure',
	'usdx',
	'xusd',
	'gho',
	'crvusd',
	'mkusd'
]);

// baseSymbol maps a NowPayments currency code (which may include a network
// suffix, e.g. "usdttrc20") to the underlying ticker used for icon lookups and
// stablecoin detection (e.g. "usdt").
export function baseSymbol(code: string): string {
	const lower = (code || '').toLowerCase().trim();
	if (!lower) return '';
	if (BASE_SYMBOL_OVERRIDES[lower]) return BASE_SYMBOL_OVERRIDES[lower];
	for (const suffix of NETWORK_SUFFIXES) {
		if (lower.endsWith(suffix)) {
			const candidate = lower.slice(0, -suffix.length);
			if (candidate.length >= 2) return candidate;
		}
	}
	return lower;
}

export function isStablecoin(code: string): boolean {
	return STABLECOINS.has(baseSymbol(code));
}

export function coinLabel(code: string): string {
	return baseSymbol(code).toUpperCase();
}
