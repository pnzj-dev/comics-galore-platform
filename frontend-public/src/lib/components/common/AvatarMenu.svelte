<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import { modal } from '$lib/stores/modal.svelte';
	import { Avatar, AvatarImage, AvatarFallback } from '$lib/components/ui/avatar/index.js';
	import {
		DropdownMenu,
		DropdownMenuTrigger,
		DropdownMenuContent,
		DropdownMenuItem,
		DropdownMenuSeparator,
	} from '$lib/components/ui/dropdown-menu/index.js';
	import { t } from '$lib/i18n';
	import { ChevronDown, User, Settings, Shield, LogOut } from 'lucide-svelte';

	const MEDIA_BASE = import.meta.env.VITE_API_URL || 'http://localhost:4000';

	const user = $derived($currentUser);
	const avatarSrc = $derived(user?.avatar_key ? `${MEDIA_BASE}/media/${user.avatar_key}` : '');
	const initial = $derived((user?.username || user?.email || '?').charAt(0).toUpperCase());

	const tierColors: Record<string, string> = {
		free: 'bg-gray-400',
		bronze: 'bg-amber-700',
		silver: 'bg-slate-400',
		gold: 'bg-amber-400',
		platinum: 'bg-cyan-400',
	};
	const tierColor = $derived(tierColors[user?.tier ?? 'free'] ?? 'bg-gray-400');

	function capitalize(s: string): string {
		return s.charAt(0).toUpperCase() + s.slice(1);
	}

	const planLabel = $derived(user?.tier && user.tier !== 'free' ? `${capitalize(user.tier)} plan` : 'Free plan');
	const roleLabel = $derived(user?.role && user.role !== 'user' ? ` · ${capitalize(user.role)}` : '');
</script>

<DropdownMenu>
	<DropdownMenuTrigger class="flex items-center gap-1.5 rounded-full p-0.5 hover:bg-muted transition-colors outline-none">
		<span class="relative">
			<Avatar class="size-8">
				{#if avatarSrc}
					<AvatarImage src={avatarSrc} alt={user?.username || user?.email || ''} />
				{/if}
				<AvatarFallback>{initial}</AvatarFallback>
			</Avatar>
			<span class="absolute -bottom-0.5 -right-0.5 size-3 rounded-full border-2 border-background {tierColor}" aria-hidden="true"></span>
		</span>
		<ChevronDown class="size-3.5 text-muted-foreground" />
	</DropdownMenuTrigger>

	<DropdownMenuContent align="end" class="w-56">
		<div class="px-2 py-1.5">
			<p class="truncate text-sm font-medium">{user?.username || user?.email}</p>
			<p class="truncate text-xs text-muted-foreground">{planLabel}{roleLabel}</p>
		</div>
		<DropdownMenuSeparator />
		<DropdownMenuItem onclick={() => modal.open('profile')}>
			<User class="size-4" />
			{t('menu.profile')}
		</DropdownMenuItem>
		<DropdownMenuItem onclick={() => modal.open('settings')}>
			<Settings class="size-4" />
			{t('menu.preferences')}
		</DropdownMenuItem>
		<DropdownMenuItem onclick={() => modal.open('security')}>
			<Shield class="size-4" />
			{t('menu.security')}
		</DropdownMenuItem>
		<DropdownMenuSeparator />
		<DropdownMenuItem variant="destructive" onclick={() => modal.open('logout')}>
			<LogOut class="size-4" />
			{t('nav.logout')}
		</DropdownMenuItem>
	</DropdownMenuContent>
</DropdownMenu>
