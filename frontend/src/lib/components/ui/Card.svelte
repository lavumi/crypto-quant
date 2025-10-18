<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		children?: Snippet;
	}

	let { class: className = '', children, ...rest }: Props = $props();
</script>

<div
	class="rounded-lg border bg-card text-card-foreground shadow-sm {className}"
	{...rest}
>
	{#if children}
		{@render children()}
	{/if}
</div>

