<!--
  AdBanner.svelte
  Homepage advertisement slot. Renders nothing unless `enabled` is true.
  - type "direct": styled banner (image / title / subtitle / CTA).
  - type "programmatic": injects a raw ad-network snippet (HTML + scripts).

  Usage:
    <AdBanner
      enabled={ad.enabled}
      type={ad.type}
      title={ad.title}
      subtitle={ad.subtitle}
      ctaText={ad.cta_text}
      ctaHref={ad.cta_href}
      imageUrl={ad.image_url}
      embedHtml={ad.embed_html}
    />
-->
<script lang="ts">
	import { onMount } from 'svelte';

	interface Props {
		enabled?: boolean;
		type?: string;
		imageUrl?: string;
		title?: string;
		subtitle?: string;
		ctaText?: string;
		ctaHref?: string;
		embedHtml?: string;
		class?: string;
	}

	let {
		enabled = false,
		type = 'direct',
		imageUrl = '',
		title = '',
		subtitle = '',
		ctaText = '',
		ctaHref = '#',
		embedHtml = '',
		class: className = ''
	}: Props = $props();

	// Programmatic ad: set the raw snippet and re-create any <script> tags so
	// they actually execute (innerHTML-inserted scripts do not run on their own).
	let embedEl = $state<HTMLDivElement>();
	onMount(() => {
		if (type !== 'programmatic' || !embedEl || !embedHtml) return;
		embedEl.innerHTML = embedHtml;
		for (const old of Array.from(embedEl.querySelectorAll('script'))) {
			const s = document.createElement('script');
			for (const attr of Array.from(old.attributes)) s.setAttribute(attr.name, attr.value);
			s.text = old.textContent || '';
			old.replaceWith(s);
		}
	});
</script>

{#if enabled}
	{#if type === 'programmatic'}
		<div bind:this={embedEl} class="w-full {className}"></div>
	{:else}
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
					{#if title}
						<h2 class="text-2xl font-bold tracking-tight text-white sm:text-3xl">{title}</h2>
					{/if}
				</div>

				{#if ctaText}
					<a
						href={ctaHref || '#'}
						class="inline-flex items-center justify-center rounded-full bg-white px-6 py-2.5 text-sm font-semibold text-gray-900 shadow-lg transition hover:bg-gray-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:ring-offset-purple-600"
					>
						{ctaText}
					</a>
				{/if}
			</div>
		</section>
	{/if}
{/if}
