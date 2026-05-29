<script lang="ts">
	import { onMount } from 'svelte';

	type Experiment = {
		id: string;
		type: 'single' | 'sweep' | 'walk_forward';
		strategy_name: string;
		symbol: string;
		interval: string;
		created_at: string;
		summary: Record<string, number | string | null>;
	};

	let items = $state<Experiment[]>([]);
	let error = $state('');
	let isLoading = $state(true);

	onMount(async () => {
		try {
			const response = await fetch('http://localhost:8080/api/v2/experiments');
			const payload = await response.json();
			if (!response.ok) throw new Error(payload?.error ?? payload?.message ?? 'Failed to load experiments');
			items = payload.data?.items ?? [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			isLoading = false;
		}
	});

	function percent(value: unknown) {
		return typeof value === 'number' ? `${(value * 100).toFixed(2)}%` : '-';
	}

	function number(value: unknown) {
		return typeof value === 'number' ? value.toFixed(2) : '-';
	}
</script>

<section class="space-y-5">
	<div class="border-b pb-4">
		<h1 class="text-2xl font-semibold">Experiments</h1>
		<p class="mt-1 text-sm text-muted-foreground">저장된 실험을 표 형태로 확인합니다.</p>
	</div>

	{#if isLoading}
		<p class="text-sm text-muted-foreground">Loading...</p>
	{:else if error}
		<div class="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">{error}</div>
	{:else}
		<div class="overflow-x-auto rounded-lg border">
			<table class="w-full text-left text-sm">
				<thead class="border-b text-muted-foreground">
					<tr>
						<th class="px-4 py-2 font-medium">Created</th>
						<th class="px-4 py-2 font-medium">Type</th>
						<th class="px-4 py-2 font-medium">Strategy</th>
						<th class="px-4 py-2 font-medium">Market</th>
						<th class="px-4 py-2 font-medium">Return</th>
						<th class="px-4 py-2 font-medium">Sharpe</th>
						<th class="px-4 py-2 font-medium">Risk</th>
					</tr>
				</thead>
				<tbody>
					{#each items as item}
						<tr class="border-b last:border-0">
							<td class="px-4 py-2 text-muted-foreground">{item.created_at}</td>
							<td class="px-4 py-2">{item.type}</td>
							<td class="px-4 py-2 font-medium">{item.strategy_name}</td>
							<td class="px-4 py-2">{item.symbol} · {item.interval}</td>
							<td class="px-4 py-2">{percent(item.summary.total_return ?? item.summary.best_return ?? item.summary.avg_test_return)}</td>
							<td class="px-4 py-2">{number(item.summary.sharpe_ratio ?? item.summary.best_sharpe_ratio ?? item.summary.avg_test_sharpe)}</td>
							<td class="px-4 py-2">{percent(item.summary.max_drawdown_pct ?? item.summary.best_max_drawdown_pct ?? item.summary.positive_test_ratio)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
