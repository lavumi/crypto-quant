<script lang="ts">
	import type { HTMLSelectAttributes } from 'svelte/elements';

	interface Props extends HTMLSelectAttributes {
		value?: string;
		options: Array<{ value: string; label: string }>;
	}

	let { class: className = '', value = $bindable(''), options, ...rest }: Props = $props();
</script>

<select
	bind:value
	class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 {className}"
	{...rest}
>
	{#each options as option}
		<option value={option.value}>{option.label}</option>
	{/each}
</select>

