<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { currentUser, isAuthenticated, hydrated } from '$lib/stores/auth';
	import UserProfileModal from '$lib/components/modals/UserProfileModal.svelte';
	import AppSettingsModal from '$lib/components/modals/AppSettingsModal.svelte';
	import LogoutConfirmationModal from '$lib/components/auth/LogoutConfirmationModal.svelte';
	import CheckoutModal from '$lib/components/billing/CheckoutModal.svelte';
	import BoostModal from '$lib/components/billing/BoostModal.svelte';
	import AddToListModal from '$lib/components/lists/AddToListModal.svelte';
	import NewMessageModal from '$lib/components/messages/NewMessageModal.svelte';
	import { initializeLocale } from '$lib/i18n';

	let { data, children } = $props();

	// Intentionally initial-value only — locale is resolved once server-side.
	// svelte-ignore state_referenced_locally
	initializeLocale(data.locale);

	onMount(() => {
		if (data.user) {
			currentUser.set(data.user);
			isAuthenticated.set(true);
		}
		hydrated.set(true);
	});
</script>

<svelte:head>
	<html lang={data.locale}></html>
</svelte:head>

{@render children()}

<UserProfileModal />
<AppSettingsModal />
<LogoutConfirmationModal />
<CheckoutModal />
<BoostModal />
<AddToListModal />
<NewMessageModal />
