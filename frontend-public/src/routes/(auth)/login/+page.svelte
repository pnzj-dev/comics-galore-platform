<script lang="ts">
	import { goto } from '$app/navigation';
	import { modal } from '$lib/stores/modal.svelte';
	import LoginForm from '$lib/components/auth/LoginForm.svelte';

	function switchToRegister() {
		modal.open('register');
	}

	function switchToForgotPassword() {
		modal.open('forgot-password');
	}
</script>

<svelte:head>
	<title>Login - Comics Galore</title>
</svelte:head>

<LoginForm
	onSwitchToRegister={switchToRegister}
	onForgotPassword={switchToForgotPassword}
	onSuccess={() => goto('/')}
/>
