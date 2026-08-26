<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { currentUser, isAuthenticated } from '$lib/stores/auth';
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import LogoutConfirmationModal from '$lib/components/LogoutConfirmationModal.svelte';
	import NowPaymentsLinkWizard from '$lib/components/NowPaymentsLinkWizard.svelte';
	import { LayoutDashboard, Shield, Users, CreditCard, BookOpen, Trash2, Settings, LogOut, AlertTriangle, ArrowDownToLine, ReceiptText, LifeBuoy, Sparkles, Ticket, Activity, HardDrive } from 'lucide-svelte';

	let { data, children } = $props();
	let planMatrixComplete = $state(true);
	let wizardBlocked = $state(false);

	onMount(() => {
		if (data.user) {
			currentUser.set(data.user);
			isAuthenticated.set(true);
		}
	});

	const user = $derived(data.user || $currentUser);
	const authed = $derived(!!(data.user));
	const path = $derived(page.url?.pathname ?? '');
	const isAdmin = $derived(user?.role === 'admin');

	async function checkPlanMatrix() {
		try {
			const res = await encore.tiers.PlanMatrixStatus();
			planMatrixComplete = res.complete;
			if (!res.complete && !wizardBlocked) modal.open('wizard');
		} catch { planMatrixComplete = true; }
	}

	function onWizardClose() { modal.close('wizard'); wizardBlocked = !planMatrixComplete; }

	$effect(() => { if (authed && isAdmin) checkPlanMatrix(); });

	const navItems = $derived(
		isAdmin
			? [
				{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
				{ href: '/moderation', label: 'Moderation', icon: Shield },
				{ href: '/support', label: 'Support', icon: LifeBuoy },
				{ href: '/ai', label: 'AI Moderation', icon: Sparkles },
				{ href: '/users', label: 'Users', icon: Users },
				{ href: '/subscriptions', label: 'Subscriptions', icon: CreditCard },
				{ href: '/deposits', label: 'Deposits', icon: ArrowDownToLine },
				{ href: '/payments', label: 'Payments', icon: ReceiptText },
				{ href: '/coupons', label: 'Coupons', icon: Ticket },
				{ href: '/comics', label: 'Comics', icon: BookOpen },
				{ href: '/series', label: 'Series', icon: BookOpen },
				{ href: '/comics/recycle-bin', label: 'Recycle Bin', icon: Trash2 },
				{ href: '/jobs', label: 'Background Jobs', icon: Activity },
				{ href: '/storage', label: 'Storage', icon: HardDrive },
				{ href: '/settings', label: 'Settings', icon: Settings },
			]
			: [
				{ href: '/moderation', label: 'Moderation', icon: Shield },
			]
	);

	function isActive(href: string): boolean {
		if (href === '/dashboard') return path === '/dashboard' || path === '/';
		return path.startsWith(href);
	}
</script>

{#if authed && (user?.role === 'admin' || user?.role === 'moderator') && page.url?.pathname !== '/login'}
	<div class="flex h-screen overflow-hidden">
		<!-- Sidebar -->
		<aside class="w-60 flex-shrink-0 bg-slate-900 dark:bg-slate-950 text-white flex flex-col">
			<div class="p-4 border-b border-slate-700/50">
				<a href="/dashboard" class="text-sm font-bold tracking-wide text-white">Comics Galore</a>
				<p class="text-[10px] text-slate-400 mt-0.5">Admin Panel</p>
			</div>

			<nav class="flex-1 p-3 space-y-0.5 overflow-y-auto">
				{#each navItems as item}
					<a
						href={item.href}
						class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors
							{isActive(item.href)
								? 'bg-white/10 text-white font-medium'
								: 'text-slate-300 hover:text-white hover:bg-white/5'}"
					>
						<item.icon class="size-4" />
						{item.label}
					</a>
				{/each}
			</nav>

			<div class="p-3 border-t border-slate-700/50">
				<button class="flex items-center gap-2.5 px-3 py-2 rounded-lg hover:bg-white/5 transition-colors cursor-pointer w-full text-left" onclick={() => modal.open('logout')}>
					<LogOut class="size-4 text-slate-400" />
					<span class="text-sm text-slate-300 hover:text-white">Sign out</span>
				</button>
				<p class="px-3 pt-2 text-[10px] text-slate-500 truncate">{user.email}</p>
			</div>
		</aside>

		<!-- Main content -->
		<div class="flex-1 flex flex-col overflow-y-auto">
			{#if !planMatrixComplete}
				<button onclick={() => { modal.open('wizard'); wizardBlocked = true; }} class="w-full flex items-center justify-center gap-1.5 bg-red-600 hover:bg-red-700 text-white text-xs text-center py-1.5 font-medium cursor-pointer border-0 transition-colors flex-shrink-0">
					<AlertTriangle class="size-3" />
					Unlinked NowPayments plans — Click here to configure
				</button>
			{/if}

			<main class="flex-1 p-6">
				{@render children()}
			</main>
		</div>
	</div>

	<LogoutConfirmationModal />
	<NowPaymentsLinkWizard onClose={onWizardClose} />
{:else}
	{@render children()}
{/if}
