<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import Card from '$lib/components/ui/Card.svelte';

	type Kind = 'single' | 'sweeps' | 'walk-forward';

	let data = $state<Record<string, any> | null>(null);
	let isLoading = $state(true);
	let error = $state('');

	function endpoint(kind: Kind, id: string) {
		if (kind === 'single') return `/api/v2/experiments/single/${id}`;
		if (kind === 'sweeps') return `/api/v2/experiments/sweeps/${id}`;
		return `/api/v2/experiments/walk-forward/${id}`;
	}

	function title(kind: Kind) {
		if (kind === 'single') return 'Single Experiment';
		if (kind === 'sweeps') return 'Sweep Experiment';
		return 'Walk-Forward Experiment';
	}

	function typeTone(kind: Kind) {
		if (kind === 'single') return 'bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-200';
		if (kind === 'sweeps') return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-200';
		return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-200';
	}

	function formatPercent(value: unknown) {
		return typeof value === 'number' ? `${(value * 100).toFixed(2)}%` : '-';
	}

	function formatNumber(value: unknown, digits = 2) {
		return typeof value === 'number' ? value.toFixed(digits) : '-';
	}

	function formatTimestamp(value: unknown) {
		if (typeof value !== 'number') return '-';
		return new Date(value * 1000).toLocaleString('ko-KR');
	}

	function summaryEntries(summary: Record<string, any> | undefined) {
		if (!summary) return [];
		return Object.entries(summary).filter(([, value]) => typeof value !== 'object' || value === null);
	}

	function metricValue(key: string, value: unknown) {
		if (typeof value !== 'number') return String(value ?? '-');
		if (key.includes('return') || key.includes('pct') || key.includes('ratio')) return formatPercent(value);
		return formatNumber(value);
	}

	function compactJson(value: unknown) {
		return JSON.stringify(value, null, 2);
	}

	onMount(async () => {
		const kind = page.params.kind as Kind;
		const id = page.params.id;
		if (!id) {
			error = 'Experiment identifier is missing';
			isLoading = false;
			return;
		}

		try {
			const response = await fetch(endpoint(kind, id));
			if (!response.ok) throw new Error(`Failed to load experiment (${response.status})`);
			const payload = await response.json();
			data = payload.data ?? null;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load experiment';
		} finally {
			isLoading = false;
		}
	});
</script>

<div class="container mx-auto max-w-7xl p-4 md:p-8">
	<div class="mb-8 rounded-3xl border bg-card p-6 shadow-sm md:p-8">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="space-y-3">
				<a href="/experiments" class="text-primary hover:underline text-sm">← experiments</a>
				<div class="flex flex-wrap items-center gap-2">
					<span class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] ${typeTone(page.params.kind as Kind)}`}>
						{page.params.kind}
					</span>
					{#if data}
						<span class="text-sm text-muted-foreground">{data.created_at}</span>
					{/if}
				</div>
				<h1 class="text-3xl font-bold md:text-4xl">{title(page.params.kind as Kind)}</h1>
				{#if data}
					<p class="text-muted-foreground">{data.strategy_name} · {data.symbol} · {data.interval}</p>
				{/if}
			</div>
			{#if data}
				<div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
					<div class="rounded-2xl border bg-background px-4 py-3">
						<p class="text-xs uppercase tracking-wide text-muted-foreground">Type</p>
						<p class="mt-1 font-semibold">{data.type}</p>
					</div>
					<div class="rounded-2xl border bg-background px-4 py-3">
						<p class="text-xs uppercase tracking-wide text-muted-foreground">Identifier</p>
						<p class="mt-1 truncate font-mono text-sm">{data.id}</p>
					</div>
					<div class="rounded-2xl border bg-background px-4 py-3">
						<p class="text-xs uppercase tracking-wide text-muted-foreground">Focus</p>
						<p class="mt-1 font-semibold">{data.selection_metric ?? data.sort_by ?? 'run'}</p>
					</div>
				</div>
			{/if}
		</div>
	</div>

	{#if isLoading}
		<Card class="p-6">
			<p class="text-muted-foreground">실험 상세를 불러오는 중...</p>
		</Card>
	{:else if error}
		<Card class="border-destructive/30 bg-destructive/5 p-6">
			<p class="font-medium text-destructive">{error}</p>
		</Card>
	{:else if data}
		<div class="space-y-6">
			{#if data.summary}
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
					{#each summaryEntries(data.summary).slice(0, 8) as [key, value]}
						<Card class="p-5">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">{key}</p>
							<p class="mt-3 text-2xl font-semibold">{metricValue(key, value)}</p>
						</Card>
					{/each}
				</div>
			{/if}

			<div class="grid grid-cols-1 gap-6 xl:grid-cols-[1.2fr_0.8fr]">
				{#if data.results}
					<Card class="p-6">
						<div class="mb-5 flex items-end justify-between gap-4">
							<div>
								<h2 class="text-xl font-semibold">Ranked Candidates</h2>
								<p class="mt-1 text-sm text-muted-foreground">상위 파라미터 조합과 주요 지표를 빠르게 비교합니다.</p>
							</div>
						</div>
						<div class="overflow-x-auto">
							<table class="w-full text-left text-sm">
								<thead class="text-muted-foreground">
									<tr class="border-b">
										<th class="px-3 py-2">Rank</th>
										<th class="px-3 py-2">Return</th>
										<th class="px-3 py-2">Sharpe</th>
										<th class="px-3 py-2">Calmar</th>
										<th class="px-3 py-2">MDD</th>
										<th class="px-3 py-2">PF</th>
										<th class="px-3 py-2">Trades</th>
									</tr>
								</thead>
								<tbody>
									{#each data.results as row}
										<tr class="border-b last:border-b-0">
											<td class="px-3 py-2 font-medium">{row.rank}</td>
											<td class="px-3 py-2">{formatPercent(row.total_return)}</td>
											<td class="px-3 py-2">{formatNumber(row.sharpe_ratio)}</td>
											<td class="px-3 py-2">{formatNumber(row.calmar_ratio)}</td>
											<td class="px-3 py-2">{formatPercent(row.max_drawdown_pct)}</td>
											<td class="px-3 py-2">{formatNumber(row.profit_factor)}</td>
											<td class="px-3 py-2">{row.total_trades ?? '-'}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</Card>
				{:else if data.windows}
					<Card class="p-6">
						<div class="mb-5">
							<h2 class="text-xl font-semibold">Walk-Forward Windows</h2>
							<p class="mt-1 text-sm text-muted-foreground">window 별 train 선택과 out-of-sample 성과를 추적합니다.</p>
						</div>
						<div class="overflow-x-auto">
							<table class="w-full text-left text-sm">
								<thead class="text-muted-foreground">
									<tr class="border-b">
										<th class="px-3 py-2">Window</th>
										<th class="px-3 py-2">Train</th>
										<th class="px-3 py-2">Test</th>
										<th class="px-3 py-2">Train Sharpe</th>
										<th class="px-3 py-2">Test Return</th>
										<th class="px-3 py-2">Test Sharpe</th>
										<th class="px-3 py-2">Test MDD</th>
									</tr>
								</thead>
								<tbody>
									{#each data.windows as row}
										<tr class="border-b last:border-b-0 align-top">
											<td class="px-3 py-2 font-medium">{row.window_index}</td>
											<td class="px-3 py-2">
												{formatTimestamp(row.train_start)}<br />
												<span class="text-muted-foreground">{formatTimestamp(row.train_end)}</span>
											</td>
											<td class="px-3 py-2">
												{formatTimestamp(row.test_start)}<br />
												<span class="text-muted-foreground">{formatTimestamp(row.test_end)}</span>
											</td>
											<td class="px-3 py-2">{formatNumber(row.train_sharpe)}</td>
											<td class="px-3 py-2">{formatPercent(row.test_return)}</td>
											<td class="px-3 py-2">{formatNumber(row.test_sharpe)}</td>
											<td class="px-3 py-2">{formatPercent(row.test_max_drawdown_pct)}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</Card>
				{:else}
					<Card class="p-6">
						<div class="mb-5">
							<h2 class="text-xl font-semibold">Run Detail</h2>
							<p class="mt-1 text-sm text-muted-foreground">저장된 단일 백테스트 실행 결과를 요약해 보여줍니다.</p>
						</div>
						{#if data.equity_curve}
							<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
								<div class="rounded-2xl border bg-background p-4">
									<p class="text-xs uppercase tracking-wide text-muted-foreground">Equity Points</p>
									<p class="mt-2 text-2xl font-semibold">{data.equity_curve.length}</p>
								</div>
								<div class="rounded-2xl border bg-background p-4">
									<p class="text-xs uppercase tracking-wide text-muted-foreground">Trades</p>
									<p class="mt-2 text-2xl font-semibold">{data.trades?.length ?? 0}</p>
								</div>
								<div class="rounded-2xl border bg-background p-4">
									<p class="text-xs uppercase tracking-wide text-muted-foreground">Time Span</p>
									<p class="mt-2 text-sm font-semibold">{formatTimestamp(data.summary?.start_time)}<br />{formatTimestamp(data.summary?.end_time)}</p>
								</div>
							</div>
						{/if}
					</Card>
				{/if}

				<div class="space-y-6">
					{#if data.config || data.base_config}
						<Card class="p-6">
							<h2 class="mb-4 text-xl font-semibold">Configuration</h2>
							<pre class="overflow-x-auto rounded-2xl border bg-background p-4 text-sm">{compactJson(data.config ?? data.base_config)}</pre>
						</Card>
					{/if}

					{#if data.parameter_grid}
						<Card class="p-6">
							<h2 class="mb-4 text-xl font-semibold">Parameter Grid</h2>
							<pre class="overflow-x-auto rounded-2xl border bg-background p-4 text-sm">{compactJson(data.parameter_grid)}</pre>
						</Card>
					{/if}

					{#if data.window_config}
						<Card class="p-6">
							<h2 class="mb-4 text-xl font-semibold">Window Config</h2>
							<pre class="overflow-x-auto rounded-2xl border bg-background p-4 text-sm">{compactJson(data.window_config)}</pre>
						</Card>
					{/if}
				</div>
			</div>

			{#if data.trades}
				<Card class="p-6">
					<div class="mb-5">
						<h2 class="text-xl font-semibold">Trades</h2>
						<p class="mt-1 text-sm text-muted-foreground">최대 100개까지 표시합니다.</p>
					</div>
					<div class="overflow-x-auto">
						<table class="w-full text-left text-sm">
							<thead class="text-muted-foreground">
								<tr class="border-b">
									<th class="px-3 py-2">Time</th>
									<th class="px-3 py-2">Side</th>
									<th class="px-3 py-2">Price</th>
									<th class="px-3 py-2">Qty</th>
									<th class="px-3 py-2">Fee</th>
									<th class="px-3 py-2">Balance</th>
								</tr>
							</thead>
							<tbody>
								{#each data.trades.slice(0, 100) as row}
									<tr class="border-b last:border-b-0">
										<td class="px-3 py-2">{formatTimestamp(row.timestamp)}</td>
										<td class="px-3 py-2">{row.side}</td>
										<td class="px-3 py-2">{formatNumber(row.price)}</td>
										<td class="px-3 py-2">{formatNumber(row.quantity, 6)}</td>
										<td class="px-3 py-2">{formatNumber(row.fee)}</td>
										<td class="px-3 py-2">{formatNumber(row.balance)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</Card>
			{/if}

			{#if data.equity_curve}
				<Card class="p-6">
					<div class="mb-5">
						<h2 class="text-xl font-semibold">Equity Curve Sample</h2>
						<p class="mt-1 text-sm text-muted-foreground">테이블로 먼저 노출하고, 다음 단계에서 차트로 확장합니다.</p>
					</div>
					<div class="overflow-x-auto">
						<table class="w-full text-left text-sm">
							<thead class="text-muted-foreground">
								<tr class="border-b">
									<th class="px-3 py-2">Time</th>
									<th class="px-3 py-2">Equity</th>
									<th class="px-3 py-2">Price</th>
								</tr>
							</thead>
							<tbody>
								{#each data.equity_curve.slice(0, 50) as row}
									<tr class="border-b last:border-b-0">
										<td class="px-3 py-2">{formatTimestamp(row.timestamp)}</td>
										<td class="px-3 py-2">{formatNumber(row.equity)}</td>
										<td class="px-3 py-2">{formatNumber(row.price)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</Card>
			{/if}
		</div>
	{/if}
</div>
