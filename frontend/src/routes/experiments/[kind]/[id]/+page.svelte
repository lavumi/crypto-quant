<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/state';
	import * as echarts from 'echarts';
	import Card from '$lib/components/ui/Card.svelte';

	type Kind = 'single' | 'sweeps' | 'walk-forward';

	let data = $state<Record<string, any> | null>(null);
	let isLoading = $state(true);
	let error = $state('');
	let primaryChartContainer = $state<HTMLDivElement | undefined>();
	let secondaryChartContainer = $state<HTMLDivElement | undefined>();
	let primaryChart: echarts.ECharts | null = null;
	let secondaryChart: echarts.ECharts | null = null;

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

	function asNumber(value: unknown) {
		return typeof value === 'number' ? value : null;
	}

	function formatPercent(value: unknown) {
		const numeric = asNumber(value);
		return numeric !== null ? `${(numeric * 100).toFixed(2)}%` : '-';
	}

	function formatNumber(value: unknown, digits = 2) {
		const numeric = asNumber(value);
		return numeric !== null ? numeric.toFixed(digits) : '-';
	}

	function formatTimestamp(value: unknown) {
		if (typeof value !== 'number') return '-';
		return new Date(value * 1000).toLocaleString('ko-KR');
	}

	function summaryEntries(summary: Record<string, any> | undefined) {
		if (!summary) return [];
		return Object.entries(summary).filter(([, value]) => typeof value !== 'object' || value === null);
	}

	function isPercentMetric(key: string) {
		return (
			key.includes('return') ||
			key.includes('drawdown') ||
			key.includes('win_rate') ||
			key.includes('positive') ||
			key.includes('_pct')
		);
	}

	function metricValue(key: string, value: unknown) {
		if (typeof value !== 'number') return String(value ?? '-');
		if (isPercentMetric(key)) return formatPercent(value);
		return formatNumber(value);
	}

	function compactJson(value: unknown) {
		return JSON.stringify(value, null, 2);
	}

	function summaryCards(kind: Kind, payload: Record<string, any>) {
		if (kind === 'single') {
			return [
				{ label: 'Return', value: formatPercent(payload.summary?.total_return) },
				{ label: 'Sharpe', value: formatNumber(payload.summary?.sharpe_ratio) },
				{ label: 'Max Drawdown', value: formatPercent(payload.summary?.max_drawdown_pct) },
				{ label: 'Trades', value: String(payload.summary?.total_trades ?? '-') }
			];
		}

		if (kind === 'sweeps') {
			return [
				{ label: 'Candidates', value: String(payload.summary?.total_candidates ?? '-') },
				{ label: 'Successful', value: String(payload.summary?.successful_runs ?? '-') },
				{ label: 'Best Sharpe', value: formatNumber(payload.summary?.best?.sharpe_ratio) },
				{ label: 'Median Return', value: formatPercent(payload.summary?.median?.total_return) }
			];
		}

		return [
			{ label: 'Windows', value: `${payload.summary?.completed_windows ?? '-'} / ${payload.summary?.total_windows ?? '-'}` },
			{ label: 'Avg OOS Return', value: formatPercent(payload.summary?.avg_test_return) },
			{ label: 'Avg OOS Sharpe', value: formatNumber(payload.summary?.avg_test_sharpe) },
			{ label: 'Positive Ratio', value: formatPercent(payload.summary?.positive_test_ratio) }
		];
	}

	function rankedSnapshots(payload: Record<string, any>) {
		if (!payload.summary) return [];
		return [
			{ label: 'Best', value: payload.summary.best },
			{ label: 'Median', value: payload.summary.median },
			{ label: 'Worst', value: payload.summary.worst }
		].filter((item) => item.value);
	}

	function setupPrimaryChart() {
		if (!primaryChartContainer || !data) return;
		const isDark = document.body.classList.contains('dark');
		primaryChart?.dispose();
		primaryChart = echarts.init(primaryChartContainer, isDark ? 'dark' : undefined);

		const kind = page.params.kind as Kind;
		if (kind === 'single' && data.equity_curve?.length) {
			const equityData = data.equity_curve.map((row: Record<string, number>) => [row.timestamp * 1000, row.equity]);
			const tradeMarkers = (data.trades ?? []).slice(0, 200).map((row: Record<string, unknown>) => ({
				name: String(row.side ?? 'TRADE'),
				coord: [Number(row.timestamp) * 1000, Number(row.balance ?? row.position ?? 0)],
				itemStyle: { color: row.side === 'BUY' ? '#10b981' : '#ef4444' }
			}));

			primaryChart.setOption({
				title: { text: 'Equity Curve', left: 'center', top: 12 },
				tooltip: { trigger: 'axis' },
				grid: { left: '4%', right: '4%', top: 56, bottom: 48, containLabel: true },
				xAxis: { type: 'time' },
				yAxis: { type: 'value', scale: true },
				dataZoom: [{ type: 'inside' }, { type: 'slider', height: 22, bottom: 10 }],
				series: [
					{
						name: 'Equity',
						type: 'line',
						smooth: true,
						showSymbol: false,
						lineStyle: { width: 3, color: '#0f766e' },
						areaStyle: { color: 'rgba(15,118,110,0.12)' },
						data: equityData,
						markPoint: { symbol: 'circle', symbolSize: 10, data: tradeMarkers }
					}
				]
			});
			return;
		}

		if (kind === 'sweeps' && data.results?.length) {
			const rows = data.results.slice(0, 12);
			primaryChart.setOption({
				title: { text: 'Top Candidates', left: 'center', top: 12 },
				tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
				legend: { top: 36 },
				grid: { left: '4%', right: '4%', top: 84, bottom: 48, containLabel: true },
				xAxis: {
					type: 'category',
					data: rows.map((row: Record<string, unknown>) => `#${row.rank}`)
				},
				yAxis: [
					{ type: 'value', name: 'Return' },
					{ type: 'value', name: 'Sharpe' }
				],
				series: [
					{
						name: 'Return',
						type: 'bar',
						data: rows.map((row: Record<string, unknown>) => asNumber(row.total_return)),
						itemStyle: { color: '#c2410c' }
					},
					{
						name: 'Sharpe',
						type: 'line',
						yAxisIndex: 1,
						smooth: true,
						data: rows.map((row: Record<string, unknown>) => asNumber(row.sharpe_ratio)),
						lineStyle: { width: 3, color: '#1d4ed8' }
					}
				]
			});
			return;
		}

		if (kind === 'walk-forward' && data.windows?.length) {
			primaryChart.setOption({
				title: { text: 'Out-of-Sample Windows', left: 'center', top: 12 },
				tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
				legend: { top: 36 },
				grid: { left: '4%', right: '4%', top: 84, bottom: 48, containLabel: true },
				xAxis: {
					type: 'category',
					data: data.windows.map((row: Record<string, unknown>) => `W${row.window_index}`)
				},
				yAxis: [
					{ type: 'value', name: 'Return' },
					{ type: 'value', name: 'Sharpe' }
				],
				series: [
					{
						name: 'Test Return',
						type: 'bar',
						data: data.windows.map((row: Record<string, unknown>) => asNumber(row.test_return)),
						itemStyle: {
							color: (params: { value: number }) => (params.value >= 0 ? '#15803d' : '#b91c1c')
						}
					},
					{
						name: 'Test Sharpe',
						type: 'line',
						yAxisIndex: 1,
						smooth: true,
						data: data.windows.map((row: Record<string, unknown>) => asNumber(row.test_sharpe)),
						lineStyle: { width: 3, color: '#0f766e' }
					}
				]
			});
		}
	}

	function setupSecondaryChart() {
		if (!secondaryChartContainer || !data) return;
		const isDark = document.body.classList.contains('dark');
		secondaryChart?.dispose();
		secondaryChart = echarts.init(secondaryChartContainer, isDark ? 'dark' : undefined);

		const kind = page.params.kind as Kind;
		if (kind === 'single' && data.equity_curve?.length) {
			secondaryChart.setOption({
				title: { text: 'Underlying Price', left: 'center', top: 12 },
				tooltip: { trigger: 'axis' },
				grid: { left: '4%', right: '4%', top: 56, bottom: 48, containLabel: true },
				xAxis: { type: 'time' },
				yAxis: { type: 'value', scale: true },
				dataZoom: [{ type: 'inside' }, { type: 'slider', height: 22, bottom: 10 }],
				series: [
					{
						name: 'Price',
						type: 'line',
						showSymbol: false,
						smooth: true,
						lineStyle: { width: 2, color: '#7c3aed' },
						data: data.equity_curve.map((row: Record<string, number>) => [row.timestamp * 1000, row.price])
					}
				]
			});
			return;
		}

		if (kind === 'sweeps' && data.results?.length) {
			const rows = data.results.slice(0, 12);
			secondaryChart.setOption({
				title: { text: 'Risk Profile', left: 'center', top: 12 },
				tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
				legend: { top: 36 },
				grid: { left: '4%', right: '4%', top: 84, bottom: 48, containLabel: true },
				xAxis: {
					type: 'category',
					data: rows.map((row: Record<string, unknown>) => `#${row.rank}`)
				},
				yAxis: { type: 'value', scale: true },
				series: [
					{
						name: 'Max Drawdown',
						type: 'bar',
						data: rows.map((row: Record<string, unknown>) => asNumber(row.max_drawdown_pct)),
						itemStyle: { color: '#dc2626' }
					},
					{
						name: 'Profit Factor',
						type: 'line',
						smooth: true,
						data: rows.map((row: Record<string, unknown>) => asNumber(row.profit_factor)),
						lineStyle: { width: 3, color: '#2563eb' }
					}
				]
			});
			return;
		}

		if (kind === 'walk-forward' && data.windows?.length) {
			secondaryChart.setOption({
				title: { text: 'Window Risk', left: 'center', top: 12 },
				tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
				grid: { left: '4%', right: '4%', top: 56, bottom: 48, containLabel: true },
				xAxis: {
					type: 'category',
					data: data.windows.map((row: Record<string, unknown>) => `W${row.window_index}`)
				},
				yAxis: { type: 'value', scale: true },
				series: [
					{
						name: 'Test Max Drawdown',
						type: 'bar',
						data: data.windows.map((row: Record<string, unknown>) => asNumber(row.test_max_drawdown_pct)),
						itemStyle: { color: '#b91c1c' }
					}
				]
			});
		}
	}

	async function refreshCharts() {
		await tick();
		setupPrimaryChart();
		setupSecondaryChart();
	}

	onMount(() => {
		const kind = page.params.kind as Kind;
		const id = page.params.id;
		async function loadExperiment() {
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
				await refreshCharts();
			} catch (err) {
				error = err instanceof Error ? err.message : 'Failed to load experiment';
			} finally {
				isLoading = false;
			}
		}

		loadExperiment();

		const handleResize = () => {
			primaryChart?.resize();
			secondaryChart?.resize();
		};

		window.addEventListener('resize', handleResize);
		return () => {
			window.removeEventListener('resize', handleResize);
			primaryChart?.dispose();
			secondaryChart?.dispose();
		};
	});

	$effect(() => {
		if (data) {
			refreshCharts();
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
					{#each summaryCards(page.params.kind as Kind, data) as card}
						<Card class="p-5">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">{card.label}</p>
							<p class="mt-3 text-2xl font-semibold">{card.value}</p>
						</Card>
					{/each}
				</div>
			{/if}

			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<Card class="p-6">
					<div class="mb-4">
						<h2 class="text-xl font-semibold">Performance View</h2>
						<p class="mt-1 text-sm text-muted-foreground">
							{page.params.kind === 'single'
								? '자산 곡선과 체결 포인트를 함께 확인합니다.'
								: page.params.kind === 'sweeps'
									? '상위 후보의 수익성과 샤프를 한 화면에서 비교합니다.'
									: 'window 별 out-of-sample 성과를 시계열로 확인합니다.'}
						</p>
					</div>
					<div bind:this={primaryChartContainer} class="h-[360px] w-full rounded-2xl border bg-background"></div>
				</Card>

				<Card class="p-6">
					<div class="mb-4">
						<h2 class="text-xl font-semibold">Risk View</h2>
						<p class="mt-1 text-sm text-muted-foreground">
							{page.params.kind === 'single'
								? '기초 가격 흐름을 함께 보며 성과 구간을 해석합니다.'
								: page.params.kind === 'sweeps'
									? 'drawdown 과 profit factor 관점에서 조합을 압축해서 봅니다.'
									: 'window 별 drawdown 패턴을 빠르게 점검합니다.'}
						</p>
					</div>
					<div bind:this={secondaryChartContainer} class="h-[360px] w-full rounded-2xl border bg-background"></div>
				</Card>
			</div>

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
					<div class="space-y-6">
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

						{#if data.summary}
							<Card class="p-6">
								<h2 class="mb-4 text-xl font-semibold">Metrics Snapshot</h2>
								<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
									{#each summaryEntries(data.summary).slice(0, 10) as [key, value]}
										<div class="rounded-2xl border bg-background px-4 py-3">
											<p class="text-xs uppercase tracking-wide text-muted-foreground">{key}</p>
											<p class="mt-2 text-lg font-semibold">{metricValue(key, value)}</p>
										</div>
									{/each}
								</div>
							</Card>
						{/if}
					</div>
				{/if}

				<div class="space-y-6">
					{#if data.results && rankedSnapshots(data).length > 0}
						<Card class="p-6">
							<h2 class="mb-4 text-xl font-semibold">Best / Median / Worst</h2>
							<div class="grid grid-cols-1 gap-3">
								{#each rankedSnapshots(data) as snapshot}
									<div class="rounded-2xl border bg-background p-4">
										<div class="flex items-center justify-between gap-4">
											<p class="text-sm font-semibold">{snapshot.label}</p>
											<p class="text-xs uppercase tracking-wide text-muted-foreground">Rank #{snapshot.value.rank}</p>
										</div>
										<div class="mt-3 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
											<div>
												<p class="text-xs uppercase tracking-wide text-muted-foreground">Return</p>
												<p class="mt-1 font-semibold">{formatPercent(snapshot.value.total_return)}</p>
											</div>
											<div>
												<p class="text-xs uppercase tracking-wide text-muted-foreground">Sharpe</p>
												<p class="mt-1 font-semibold">{formatNumber(snapshot.value.sharpe_ratio)}</p>
											</div>
											<div>
												<p class="text-xs uppercase tracking-wide text-muted-foreground">MDD</p>
												<p class="mt-1 font-semibold">{formatPercent(snapshot.value.max_drawdown_pct)}</p>
											</div>
											<div>
												<p class="text-xs uppercase tracking-wide text-muted-foreground">PF</p>
												<p class="mt-1 font-semibold">{formatNumber(snapshot.value.profit_factor)}</p>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</Card>
					{/if}

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
