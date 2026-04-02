<script lang="ts">
	import { onMount } from 'svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	type ExperimentType = 'single' | 'sweep' | 'walk_forward';

	interface ExperimentItem {
		id: string;
		type: ExperimentType;
		strategy_name: string;
		symbol: string;
		interval: string;
		created_at: string;
		summary: Record<string, number | string | null>;
	}

	const filters: Array<{ key: 'all' | ExperimentType; label: string }> = [
		{ key: 'all', label: 'All' },
		{ key: 'single', label: 'Single' },
		{ key: 'sweep', label: 'Sweep' },
		{ key: 'walk_forward', label: 'Walk-Forward' }
	];

	let items = $state<ExperimentItem[]>([]);
	let isLoading = $state(true);
	let error = $state('');
	let activeFilter = $state<'all' | ExperimentType>('all');
	let query = $state('');

	function detailHref(item: ExperimentItem) {
		const rawId = item.id.split('_').pop() ?? item.id;
		const kind =
			item.type === 'walk_forward' ? 'walk-forward' : item.type === 'sweep' ? 'sweeps' : 'single';
		return `/experiments/${kind}/${rawId}`;
	}

	function formatPercent(value: unknown) {
		return typeof value === 'number' ? `${(value * 100).toFixed(2)}%` : '-';
	}

	function formatNumber(value: unknown, digits = 2) {
		return typeof value === 'number' ? value.toFixed(digits) : '-';
	}

	function headline(item: ExperimentItem) {
		if (item.type === 'single') return `Return ${formatPercent(item.summary.total_return)}`;
		if (item.type === 'sweep') return `Best Sharpe ${formatNumber(item.summary.best_sharpe_ratio)}`;
		return `Avg OOS ${formatPercent(item.summary.avg_test_return)}`;
	}

	function subheadline(item: ExperimentItem) {
		if (item.type === 'single') {
			return `Trades ${item.summary.total_trades ?? '-'} · Sharpe ${formatNumber(item.summary.sharpe_ratio)}`;
		}
		if (item.type === 'sweep') {
			return `${item.summary.total_candidates ?? '-'} candidates · ${item.summary.successful_runs ?? '-'} successful`;
		}
		return `${item.summary.completed_windows ?? '-'} / ${item.summary.total_windows ?? '-'} windows completed`;
	}

	function typeTone(type: ExperimentType) {
		if (type === 'single') return 'bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-200';
		if (type === 'sweep') return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-200';
		return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-200';
	}

	function summaryValue(item: ExperimentItem, key: 'primary' | 'secondary') {
		if (key === 'primary') {
			return item.type === 'walk_forward'
				? formatNumber(item.summary.avg_test_sharpe)
				: formatNumber(item.summary.sharpe_ratio ?? item.summary.best_sharpe_ratio);
		}

		return item.type === 'walk_forward'
			? formatPercent(item.summary.positive_test_ratio)
			: formatPercent(item.summary.max_drawdown_pct ?? item.summary.best_max_drawdown_pct);
	}

	const filteredItems = $derived.by(() => {
		const normalizedQuery = query.trim().toLowerCase();
		return items.filter((item) => {
			const matchesType = activeFilter === 'all' || item.type === activeFilter;
			const haystack = [item.strategy_name, item.symbol, item.interval, item.id, item.type].join(' ').toLowerCase();
			const matchesQuery = !normalizedQuery || haystack.includes(normalizedQuery);
			return matchesType && matchesQuery;
		});
	});

	const counts = $derived.by(() => ({
		all: items.length,
		single: items.filter((item) => item.type === 'single').length,
		sweep: items.filter((item) => item.type === 'sweep').length,
		walk_forward: items.filter((item) => item.type === 'walk_forward').length
	}));

	onMount(async () => {
		try {
			const response = await fetch('/api/v2/experiments');
			if (!response.ok) throw new Error(`Failed to load experiments (${response.status})`);
			const payload = await response.json();
			items = payload.data?.items ?? [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load experiments';
		} finally {
			isLoading = false;
		}
	});
</script>

<div class="container mx-auto max-w-7xl p-4 md:p-8">
	<div class="mb-8 rounded-3xl border bg-card p-6 shadow-sm md:p-8">
		<div class="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-2xl space-y-3">
				<a href="/" class="text-primary hover:underline text-sm">← 홈으로</a>
				<div>
					<p class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">Research Archive</p>
					<h1 class="mt-2 text-3xl font-bold md:text-4xl">Experiments</h1>
					<p class="mt-3 text-muted-foreground">
						저장된 single backtest, parameter sweep, walk-forward 실험을 한 곳에서 비교하고
						다시 열람합니다.
					</p>
				</div>
			</div>
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
				{#each filters as filter}
					<div class="rounded-2xl border bg-background px-4 py-3">
						<p class="text-xs uppercase tracking-wide text-muted-foreground">{filter.label}</p>
						<p class="mt-1 text-2xl font-semibold">
							{filter.key === 'all' ? counts.all : counts[filter.key]}
						</p>
					</div>
				{/each}
			</div>
		</div>
	</div>

	<div class="mb-6 flex flex-col gap-4 rounded-2xl border bg-card p-4 md:flex-row md:items-center md:justify-between">
		<div class="flex flex-wrap gap-2">
			{#each filters as filter}
				<Button
					variant={activeFilter === filter.key ? 'default' : 'outline'}
					size="sm"
					onclick={() => {
						activeFilter = filter.key;
					}}
				>
					{filter.label}
				</Button>
			{/each}
		</div>
		<div class="w-full md:max-w-sm">
			<input
				class="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none transition focus:ring-2 focus:ring-ring"
				placeholder="Search by strategy, symbol, interval..."
				bind:value={query}
			/>
		</div>
	</div>

	{#if isLoading}
		<Card class="p-6">
			<p class="text-muted-foreground">실험 목록을 불러오는 중...</p>
		</Card>
	{:else if error}
		<Card class="border-destructive/30 bg-destructive/5 p-6">
			<p class="font-medium text-destructive">{error}</p>
		</Card>
	{:else if filteredItems.length === 0}
		<Card class="p-8">
			<h2 class="text-lg font-semibold">No experiments found</h2>
			<p class="mt-2 text-muted-foreground">
				현재 필터에 맞는 저장 실험이 없습니다. sweep 또는 walk-forward를 한 번 실행해보세요.
			</p>
		</Card>
	{:else}
		<div class="grid grid-cols-1 gap-4">
			{#each filteredItems as item}
				<a href={detailHref(item)} class="block">
					<Card class="overflow-hidden transition-all hover:-translate-y-0.5 hover:shadow-lg">
						<div class="border-b bg-linear-to-r from-background via-background to-secondary/40 px-5 py-4 md:px-6">
							<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
								<div class="space-y-2">
									<div class="flex flex-wrap items-center gap-2">
										<span class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.16em] ${typeTone(item.type)}`}>
											{item.type}
										</span>
										<span class="text-sm text-muted-foreground">{item.created_at}</span>
									</div>
									<div>
										<h2 class="text-xl font-semibold">{item.strategy_name}</h2>
										<p class="text-sm text-muted-foreground">{item.symbol} · {item.interval}</p>
									</div>
								</div>
								<div class="max-w-sm text-sm">
									<p class="font-semibold">{headline(item)}</p>
									<p class="mt-1 text-muted-foreground">{subheadline(item)}</p>
								</div>
							</div>
						</div>

						<div class="grid grid-cols-1 gap-3 px-5 py-4 md:grid-cols-4 md:px-6">
							<div class="rounded-2xl border bg-background p-4">
								<p class="text-xs uppercase tracking-wide text-muted-foreground">Primary Metric</p>
								<p class="mt-2 text-xl font-semibold">{summaryValue(item, 'primary')}</p>
							</div>
							<div class="rounded-2xl border bg-background p-4">
								<p class="text-xs uppercase tracking-wide text-muted-foreground">Risk / Hit Ratio</p>
								<p class="mt-2 text-xl font-semibold">{summaryValue(item, 'secondary')}</p>
							</div>
							<div class="rounded-2xl border bg-background p-4">
								<p class="text-xs uppercase tracking-wide text-muted-foreground">Identifier</p>
								<p class="mt-2 truncate font-mono text-sm">{item.id}</p>
							</div>
							<div class="rounded-2xl border bg-background p-4">
								<p class="text-xs uppercase tracking-wide text-muted-foreground">Open</p>
								<p class="mt-2 text-sm font-semibold text-primary">View detail →</p>
							</div>
						</div>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</div>
