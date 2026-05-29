<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();

	const nav = [
		{ href: '/backtest/new', label: 'Backtest' },
		{ href: '/experiments', label: 'Experiments' },
		{ href: '/swagger/index.html', label: 'Swagger' }
	];
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
	<header class="border-b">
		<div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
			<a href="/" class="text-sm font-semibold tracking-wide">Crypto Quant</a>
			<nav class="flex items-center gap-1">
				{#each nav as item}
					<a class="rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground" href={item.href}>
						{item.label}
					</a>
				{/each}
			</nav>
		</div>
	</header>

	<main class="mx-auto max-w-6xl px-4 py-6">
		{@render children?.()}
	</main>
</div>
