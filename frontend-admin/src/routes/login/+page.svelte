<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
</script>

<svelte:head><title>Admin Login — Comics Galore</title></svelte:head>

<div class="flex min-h-[80vh] items-center justify-center p-4">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Admin Login</CardTitle>
			<CardDescription>Sign in to the admin panel</CardDescription>
		</CardHeader>
		<CardContent>
			<form method="POST" action="?/signIn" class="space-y-4">
				<Button type="submit" class="w-full">Sign in with Logto</Button>
			</form>
			<form method="POST" action="?/forgotPassword" class="mt-3 text-center">
				<button type="submit" class="text-sm text-muted-foreground hover:text-foreground hover:underline">Forgot password?</button>
			</form>
		</CardContent>
	</Card>
</div>
