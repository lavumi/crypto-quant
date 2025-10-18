<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { onMount } from 'svelte';

	let { children } = $props();

	// Initialize dark mode early to prevent flash
	if (typeof window !== 'undefined') {
		const savedMode = localStorage.getItem('darkMode');
		if (savedMode === 'true') {
			document.body.classList.add('dark');
		}
	}

	onMount(() => {
		// Ensure dark mode is applied after mount as well
		const savedMode = localStorage.getItem('darkMode');
		if (savedMode === 'true') {
			document.body.classList.add('dark');
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children?.()}
