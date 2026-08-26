<!--
  AdBanner.svelte
  Reserved advertising section. Real ad content is provided by an ad agency
  integration; for now it renders a placeholder.

  Usage:
    <AdBanner
      title="Your story belongs here"
      subtitle="Advertisement"
      ctaText="Learn more"
      ctaHref="#"
    />
-->
<script lang="ts">
	interface Props {
		imageUrl?: string;
		title?: string;
		subtitle?: string;
		ctaText?: string;
		ctaHref?: string;
		onCta?: () => void;
		class?: string;
	}

	let {
		imageUrl = '',
		title = 'Your story belongs here',
		subtitle = 'Advertisement',
		ctaText = 'Learn more',
		ctaHref = '#',
		onCta,
		class: className = ''
	}: Props = $props();
</script>

<section
	class="relative w-full overflow-hidden rounded-2xl bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-500 dark:from-indigo-900 dark:via-purple-900 dark:to-pink-900 {className}"
	aria-label="Advertisement"
>
	{#if imageUrl}
		<img
			src={imageUrl}
			alt=""
			class="absolute inset-0 h-full w-full object-cover opacity-40"
			loading="lazy"
		/>
	{/if}

	<div class="relative z-10 flex flex-col items-start justify-center gap-3 px-6 py-8 sm:flex-row sm:items-center sm:justify-between sm:px-10 sm:py-10">
		<div class="max-w-xl">
			{#if subtitle}
				<p class="mb-1 text-xs font-medium uppercase tracking-wider text-white/80">{subtitle}</p>
			{/if}
			<h2 class="text-2xl font-bold tracking-tight text-white sm:text-3xl">{title}</h2>
		</div>

		{#if onCta}
			<a
				href={ctaHref}
				onclick={(e) => {
					e.preventDefault();
					onCta();
				}}
				class="inline-flex items-center justify-center rounded-full bg-white px-6 py-2.5 text-sm font-semibold text-gray-900 shadow-lg transition hover:bg-gray-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:ring-offset-purple-600"
			>
				{ctaText}
			</a>
		{/if}
	</div>
</section>
