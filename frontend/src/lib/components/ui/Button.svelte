<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLButtonAttributes {
		variant?: 'default' | 'outline' | 'ghost';
		size?: 'default' | 'sm' | 'lg';
		children?: Snippet;
	}

	let { variant = 'default', size = 'default', class: className = '', children, ...rest }: Props = $props();

	const variantClasses = {
		default: 'bg-primary text-primary-foreground hover:bg-primary/90',
		outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
		ghost: 'hover:bg-accent hover:text-accent-foreground'
	};

	const sizeClasses = {
		default: 'h-10 px-4 py-2',
		sm: 'h-9 px-3 text-sm',
		lg: 'h-11 px-8'
	};
</script>

<button
	class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 {variantClasses[
		variant
	]} {sizeClasses[size]} {className}"
	{...rest}
>
	{#if children}
		{@render children()}
	{/if}
</button>

