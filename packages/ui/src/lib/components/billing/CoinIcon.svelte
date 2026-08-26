<script lang="ts">
	import Aave from 'svelte-cryptocurrency-icons/Aave.svelte';
	import Ada from 'svelte-cryptocurrency-icons/Ada.svelte';
	import Algo from 'svelte-cryptocurrency-icons/Algo.svelte';
	import Ark from 'svelte-cryptocurrency-icons/Ark.svelte';
	import Atom from 'svelte-cryptocurrency-icons/Atom.svelte';
	import Avax from 'svelte-cryptocurrency-icons/Avax.svelte';
	import Bal from 'svelte-cryptocurrency-icons/Bal.svelte';
	import Band from 'svelte-cryptocurrency-icons/Band.svelte';
	import Bat from 'svelte-cryptocurrency-icons/Bat.svelte';
	import Bch from 'svelte-cryptocurrency-icons/Bch.svelte';
	import Bnb from 'svelte-cryptocurrency-icons/Bnb.svelte';
	import Btc from 'svelte-cryptocurrency-icons/Btc.svelte';
	import Btt from 'svelte-cryptocurrency-icons/Btt.svelte';
	import Chz from 'svelte-cryptocurrency-icons/Chz.svelte';
	import Comp from 'svelte-cryptocurrency-icons/Comp.svelte';
	import Crv from 'svelte-cryptocurrency-icons/Crv.svelte';
	import Dai from 'svelte-cryptocurrency-icons/Dai.svelte';
	import Dash from 'svelte-cryptocurrency-icons/Dash.svelte';
	import Dcr from 'svelte-cryptocurrency-icons/Dcr.svelte';
	import Doge from 'svelte-cryptocurrency-icons/Doge.svelte';
	import Dot from 'svelte-cryptocurrency-icons/Dot.svelte';
	import Enj from 'svelte-cryptocurrency-icons/Enj.svelte';
	import Eos from 'svelte-cryptocurrency-icons/Eos.svelte';
	import Etc from 'svelte-cryptocurrency-icons/Etc.svelte';
	import Eth from 'svelte-cryptocurrency-icons/Eth.svelte';
	import Fil from 'svelte-cryptocurrency-icons/Fil.svelte';
	import Fun from 'svelte-cryptocurrency-icons/Fun.svelte';
	import Generic from 'svelte-cryptocurrency-icons/Generic.svelte';
	import Gno from 'svelte-cryptocurrency-icons/Gno.svelte';
	import Grt from 'svelte-cryptocurrency-icons/Grt.svelte';
	import Gusd from 'svelte-cryptocurrency-icons/Gusd.svelte';
	import Ht from 'svelte-cryptocurrency-icons/Ht.svelte';
	import Icp from 'svelte-cryptocurrency-icons/Icp.svelte';
	import Icx from 'svelte-cryptocurrency-icons/Icx.svelte';
	import Kmd from 'svelte-cryptocurrency-icons/Kmd.svelte';
	import Knc from 'svelte-cryptocurrency-icons/Knc.svelte';
	import Ksm from 'svelte-cryptocurrency-icons/Ksm.svelte';
	import Link from 'svelte-cryptocurrency-icons/Link.svelte';
	import Loom from 'svelte-cryptocurrency-icons/Loom.svelte';
	import Lrc from 'svelte-cryptocurrency-icons/Lrc.svelte';
	import Lsk from 'svelte-cryptocurrency-icons/Lsk.svelte';
	import Ltc from 'svelte-cryptocurrency-icons/Ltc.svelte';
	import Mana from 'svelte-cryptocurrency-icons/Mana.svelte';
	import Matic from 'svelte-cryptocurrency-icons/Matic.svelte';
	import Miota from 'svelte-cryptocurrency-icons/Miota.svelte';
	import Mkr from 'svelte-cryptocurrency-icons/Mkr.svelte';
	import Nano from 'svelte-cryptocurrency-icons/Nano.svelte';
	import Neo from 'svelte-cryptocurrency-icons/Neo.svelte';
	import Ok from 'svelte-cryptocurrency-icons/Ok.svelte';
	import Omg from 'svelte-cryptocurrency-icons/Omg.svelte';
	import Ont from 'svelte-cryptocurrency-icons/Ont.svelte';
	import Qnt from 'svelte-cryptocurrency-icons/Qnt.svelte';
	import Qtum from 'svelte-cryptocurrency-icons/Qtum.svelte';
	import Ren from 'svelte-cryptocurrency-icons/Ren.svelte';
	import Rep from 'svelte-cryptocurrency-icons/Rep.svelte';
	import Rvn from 'svelte-cryptocurrency-icons/Rvn.svelte';
	import Sand from 'svelte-cryptocurrency-icons/Sand.svelte';
	import Snt from 'svelte-cryptocurrency-icons/Snt.svelte';
	import Snx from 'svelte-cryptocurrency-icons/Snx.svelte';
	import Sol from 'svelte-cryptocurrency-icons/Sol.svelte';
	import Storj from 'svelte-cryptocurrency-icons/Storj.svelte';
	import Sushi from 'svelte-cryptocurrency-icons/Sushi.svelte';
	import Theta from 'svelte-cryptocurrency-icons/Theta.svelte';
	import Trx from 'svelte-cryptocurrency-icons/Trx.svelte';
	import Tusd from 'svelte-cryptocurrency-icons/Tusd.svelte';
	import Uma from 'svelte-cryptocurrency-icons/Uma.svelte';
	import Uni from 'svelte-cryptocurrency-icons/Uni.svelte';
	import Usdc from 'svelte-cryptocurrency-icons/Usdc.svelte';
	import Usdt from 'svelte-cryptocurrency-icons/Usdt.svelte';
	import Vet from 'svelte-cryptocurrency-icons/Vet.svelte';
	import Waves from 'svelte-cryptocurrency-icons/Waves.svelte';
	import Wbtc from 'svelte-cryptocurrency-icons/Wbtc.svelte';
	import Xem from 'svelte-cryptocurrency-icons/Xem.svelte';
	import Xlm from 'svelte-cryptocurrency-icons/Xlm.svelte';
	import Xmr from 'svelte-cryptocurrency-icons/Xmr.svelte';
	import Xrp from 'svelte-cryptocurrency-icons/Xrp.svelte';
	import Xtz from 'svelte-cryptocurrency-icons/Xtz.svelte';
	import Xvg from 'svelte-cryptocurrency-icons/Xvg.svelte';
	import Yfi from 'svelte-cryptocurrency-icons/Yfi.svelte';
	import Zec from 'svelte-cryptocurrency-icons/Zec.svelte';
	import Zil from 'svelte-cryptocurrency-icons/Zil.svelte';
	import Zrx from 'svelte-cryptocurrency-icons/Zrx.svelte';
	import { baseSymbol } from '$lib/utils/crypto';

	let { code, size = '24' }: { code: string; size?: string } = $props();

	const ICONS: Record<string, any> = {
		aave: Aave,
		ada: Ada,
		algo: Algo,
		ark: Ark,
		atom: Atom,
		avax: Avax,
		bal: Bal,
		band: Band,
		bat: Bat,
		bch: Bch,
		bnb: Bnb,
		btc: Btc,
		btt: Btt,
		chz: Chz,
		comp: Comp,
		crv: Crv,
		dai: Dai,
		dash: Dash,
		dcr: Dcr,
		doge: Doge,
		dot: Dot,
		enj: Enj,
		eos: Eos,
		etc: Etc,
		eth: Eth,
		fil: Fil,
		fun: Fun,
		gno: Gno,
		grt: Grt,
		gusd: Gusd,
		ht: Ht,
		icp: Icp,
		icx: Icx,
		kmd: Kmd,
		knc: Knc,
		ksm: Ksm,
		link: Link,
		loom: Loom,
		lrc: Lrc,
		lsk: Lsk,
		ltc: Ltc,
		mana: Mana,
		matic: Matic,
		miota: Miota,
		mkr: Mkr,
		nano: Nano,
		neo: Neo,
		ok: Ok,
		omg: Omg,
		ont: Ont,
		qnt: Qnt,
		qtum: Qtum,
		ren: Ren,
		rep: Rep,
		rvn: Rvn,
		sand: Sand,
		snt: Snt,
		snx: Snx,
		sol: Sol,
		storj: Storj,
		sushi: Sushi,
		theta: Theta,
		trx: Trx,
		tusd: Tusd,
		uma: Uma,
		uni: Uni,
		usdc: Usdc,
		usdt: Usdt,
		vet: Vet,
		waves: Waves,
		wbtc: Wbtc,
		xem: Xem,
		xlm: Xlm,
		xmr: Xmr,
		xrp: Xrp,
		xtz: Xtz,
		xvg: Xvg,
		yfi: Yfi,
		zec: Zec,
		zil: Zil,
		zrx: Zrx
	};

	const Icon = $derived(ICONS[baseSymbol(code)] ?? Generic);
</script>

<Icon {size} />
