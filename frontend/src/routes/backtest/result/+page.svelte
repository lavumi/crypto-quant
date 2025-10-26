<script lang="ts">
	import { onMount, tick } from 'svelte';
	import * as echarts from 'echarts';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	interface BacktestResult {
		// Time metrics
		start_time: string;
		end_time: string;
		duration: string;
		
		// Financial metrics
		initial_balance: number;
		final_equity: number;
		total_return: number;
		total_return_pct: string;
		
		// Risk metrics
		sharpe_ratio: number;
		max_drawdown: number;
		max_drawdown_pct: number;
		
		// Trade statistics
		total_trades: number;
		winning_trades: number;
		losing_trades: number;
		win_rate: number;
		win_rate_pct: string;
		
		// Configuration
		strategy: string;
		symbol: string;
		interval: string;
		commission: number;
		candles_used: number;
		
		// Trade details
		recent_trades: Array<{
			timestamp: string;
			side: 'BUY' | 'SELL';
			price: number;
			quantity: number;
			fee: number;
			balance: number;
			position: number;
			reason: string;
		}>;

		// Chart data
		chart_data?: {
			equity_curve: Array<{
				timestamp: string;
				equity: number;
				price: number;
			}>;
			trades: Array<{
				timestamp: string;
				side: 'BUY' | 'SELL';
				price: number;
				equity: number;
			}>;
			indicators?: {
				price_data: Array<{
					timestamp: string;
					price: number;
				}>;
				fast_ma?: Array<{
					timestamp: string;
					value: number;
				}>;
				slow_ma?: Array<{
					timestamp: string;
					value: number;
				}>;
				rsi?: Array<{
					timestamp: string;
					value: number;
				}>;
			};
		};
	}

	let result = $state<BacktestResult | null>(null);
	let showAllMetrics = $state(false);
	let equityChartContainer = $state<HTMLDivElement | undefined>();
	let priceChartContainer = $state<HTMLDivElement | undefined>();
	let equityChart: echarts.ECharts | null = null;
	let priceChart: echarts.ECharts | null = null;

	onMount(() => {
		const stored = sessionStorage.getItem('backtest_result');
		if (stored) {
			try {
				const parsed = JSON.parse(stored);
				result = parsed as BacktestResult;
				console.log('📊 백테스트 결과:', result);
				console.log('📊 차트 데이터:', result?.chart_data);
			} catch (e) {
				console.error('Failed to parse backtest result:', e);
			}
		}

		// Cleanup on unmount
		return () => {
			if (equityChart) {
				equityChart.dispose();
			}
			if (priceChart) {
				priceChart.dispose();
			}
		};
	});

	// Initialize equity chart when container becomes available
	$effect(() => {
		if (equityChartContainer && result?.chart_data && !equityChart) {
			console.log('🎨 자산 차트 초기화!');
			initEquityChart();
		}
	});

	// Initialize price chart when container becomes available
	$effect(() => {
		if (priceChartContainer && result?.chart_data?.indicators && !priceChart) {
			console.log('🎨 가격 차트 초기화!');
			initPriceChart();
		}
	});

	// Watch for dark mode changes
	$effect(() => {
		const isDark = document.body.classList.contains('dark');
		if (equityChart && result?.chart_data) {
			updateEquityChartTheme(isDark);
		}
		if (priceChart && result?.chart_data?.indicators) {
			updatePriceChartTheme(isDark);
		}
	});

	function initEquityChart() {
		if (!equityChartContainer || !result?.chart_data) return;

		const isDark = document.body.classList.contains('dark');
		equityChart = echarts.init(equityChartContainer, isDark ? 'dark' : undefined);
		
		updateEquityChartData();

		// Handle window resize
		window.addEventListener('resize', () => equityChart?.resize());
	}

	function initPriceChart() {
		if (!priceChartContainer || !result?.chart_data?.indicators) return;

		const isDark = document.body.classList.contains('dark');
		priceChart = echarts.init(priceChartContainer, isDark ? 'dark' : undefined);
		
		updatePriceChartData();

		// Handle window resize
		window.addEventListener('resize', () => priceChart?.resize());
	}

	function updateEquityChartTheme(isDark: boolean) {
		if (!equityChart || !equityChartContainer) return;
		
		equityChart.dispose();
		equityChart = echarts.init(equityChartContainer, isDark ? 'dark' : undefined);
		updateEquityChartData();
	}

	function updatePriceChartTheme(isDark: boolean) {
		if (!priceChart || !priceChartContainer) return;
		
		priceChart.dispose();
		priceChart = echarts.init(priceChartContainer, isDark ? 'dark' : undefined);
		updatePriceChartData();
	}

	function updateEquityChartData() {
		if (!equityChart || !result?.chart_data) return;

		const { equity_curve, trades } = result.chart_data;

		// Prepare equity curve data
		const equityData = equity_curve.map(p => [new Date(p.timestamp).getTime(), p.equity]);

		// Prepare trade markers
		const buyTrades = trades
			.filter(t => t.side === 'BUY')
			.map(t => ({
				name: 'BUY',
				coord: [new Date(t.timestamp).getTime(), t.equity],
				value: 'BUY',
				itemStyle: { color: '#10b981' }
			}));

		const sellTrades = trades
			.filter(t => t.side === 'SELL')
			.map(t => ({
				name: 'SELL',
				coord: [new Date(t.timestamp).getTime(), t.equity],
				value: 'SELL',
				itemStyle: { color: '#ef4444' }
			}));

		const option: echarts.EChartsOption = {
			title: {
				text: '포트폴리오 자산 곡선',
				left: 'center',
				top: 10
			},
			tooltip: {
				trigger: 'axis',
				axisPointer: {
					type: 'cross'
				},
				formatter: function(params: any) {
					const data = params[0];
					const date = new Date(data.value[0]).toLocaleString('ko-KR');
					const equity = data.value[1].toLocaleString('en-US', { 
						minimumFractionDigits: 2, 
						maximumFractionDigits: 2 
					});
					return `${date}<br/>자산: $${equity}`;
				}
			},
			dataZoom: [
				{
					type: 'inside',
					start: 0,
					end: 100,
					zoomOnMouseWheel: true,
					moveOnMouseMove: true,
					moveOnMouseWheel: false
				},
				{
					type: 'slider',
					start: 0,
					end: 100,
					height: 30,
					bottom: 10
				}
			],
			grid: {
				left: '3%',
				right: '4%',
				bottom: 60,
				top: 80,
				containLabel: true
			},
			xAxis: {
				type: 'time',
				axisLabel: {
					formatter: (value: number) => {
						const date = new Date(value);
						return `${date.getMonth() + 1}/${date.getDate()}`;
					}
				}
			},
			yAxis: {
				type: 'value',
				scale: true,
				axisLabel: {
					formatter: (value: number) => `$${(value / 1000).toFixed(1)}k`
				},
				splitLine: {
					show: true,
					lineStyle: {
						type: 'dashed',
						opacity: 0.3
					}
				}
			},
			series: [
				{
					name: '자산',
					type: 'line',
					data: equityData,
					smooth: true,
					symbol: 'none',
					lineStyle: {
						width: 2,
						color: '#667eea'
					},
					areaStyle: {
						color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
							{ offset: 0, color: 'rgba(102, 126, 234, 0.3)' },
							{ offset: 1, color: 'rgba(102, 126, 234, 0.05)' }
						])
					},
					markPoint: {
						symbol: 'circle',
						symbolSize: 0,
						label: {
							show: false
						},
						data: [...buyTrades, ...sellTrades]
					},
					markLine: {
						silent: true,
						lineStyle: {
							color: '#999',
							type: 'dashed'
						},
						data: [
							{
								yAxis: result.initial_balance,
								label: {
									formatter: '시작: ${c}',
									position: 'end'
								}
							}
						]
					}
				}
			]
		};

		equityChart.setOption(option);
	}

	function updatePriceChartData() {
		if (!priceChart || !result?.chart_data?.indicators) return;

		const { price_data, fast_ma, slow_ma } = result.chart_data.indicators;
		const { trades } = result.chart_data;

		// Prepare price data
		const priceData = price_data.map(p => [new Date(p.timestamp).getTime(), p.price]);

		// Prepare MA data
		const fastMAData = fast_ma?.map(p => [new Date(p.timestamp).getTime(), p.value]) || [];
		const slowMAData = slow_ma?.map(p => [new Date(p.timestamp).getTime(), p.value]) || [];

		// Prepare trade markers
		const buyTrades = trades
			.filter(t => t.side === 'BUY')
			.map(t => ({
				name: 'BUY',
				coord: [new Date(t.timestamp).getTime(), t.price],
				value: 'BUY',
				itemStyle: { color: '#10b981' }
			}));

		const sellTrades = trades
			.filter(t => t.side === 'SELL')
			.map(t => ({
				name: 'SELL',
				coord: [new Date(t.timestamp).getTime(), t.price],
				value: 'SELL',
				itemStyle: { color: '#ef4444' }
			}));

		const series: any[] = [
			{
				name: '가격',
				type: 'line',
				data: priceData,
				smooth: false,
				symbol: 'none',
				lineStyle: {
					width: 1,
					color: '#6b7280'
				},
				markPoint: {
					symbol: 'circle',
					symbolSize: 8,
					label: {
						show: false
					},
					data: [...buyTrades, ...sellTrades]
				}
			}
		];

		// Add MA lines if available
		if (fastMAData.length > 0) {
			series.push({
				name: '단기 MA',
				type: 'line',
				data: fastMAData,
				smooth: true,
				symbol: 'none',
				lineStyle: {
					width: 2,
					color: '#3b82f6'
				}
			});
		}

		if (slowMAData.length > 0) {
			series.push({
				name: '장기 MA',
				type: 'line',
				data: slowMAData,
				smooth: true,
				symbol: 'none',
				lineStyle: {
					width: 2,
					color: '#f59e0b'
				}
			});
		}

		const option: echarts.EChartsOption = {
			title: {
				text: '가격 & 이평선',
				left: 'center',
				top: 10
			},
			tooltip: {
				trigger: 'axis',
				axisPointer: {
					type: 'cross'
				}
			},
			legend: {
				data: ['가격', '단기 MA', '장기 MA'],
				top: 40,
				left: 'center'
			},
			dataZoom: [
				{
					type: 'inside',
					start: 0,
					end: 100,
					zoomOnMouseWheel: true,
					moveOnMouseMove: true,
					moveOnMouseWheel: false
				},
				{
					type: 'slider',
					start: 0,
					end: 100,
					height: 30,
					bottom: 10
				}
			],
			grid: {
				left: '3%',
				right: '4%',
				bottom: 60,
				top: 80,
				containLabel: true
			},
			xAxis: {
				type: 'time',
				axisLabel: {
					formatter: (value: number) => {
						const date = new Date(value);
						return `${date.getMonth() + 1}/${date.getDate()}`;
					}
				}
			},
			yAxis: {
				type: 'value',
				scale: true,
				axisLabel: {
					formatter: (value: number) => `$${(value / 1000).toFixed(1)}k`
				},
				splitLine: {
					show: true,
					lineStyle: {
						type: 'dashed',
						opacity: 0.3
					}
				}
			},
			series: series
		};

		priceChart.setOption(option);
	}

	function formatPercent(value?: number): string {
		if (value === undefined || value === null) return 'N/A';
		return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
	}

	function formatNumber(value?: number): string {
		if (value === undefined || value === null) return 'N/A';
		return value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function formatDate(timestamp: string): string {
		return new Date(timestamp).toLocaleString('ko-KR', {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<div class="container mx-auto p-4 md:p-8 max-w-7xl">
	{#if !result}
		<div class="flex items-center justify-center min-h-[60vh]">
			<Card class="p-8 text-center">
				<p class="text-lg text-muted-foreground mb-4">백테스트 결과가 없습니다</p>
				<Button onclick={() => (window.location.href = '/backtest/new')}>
					새 백테스트 실행
				</Button>
			</Card>
		</div>
	{:else}
		<!-- Header -->
		<div class="mb-8">
			<a href="/backtest/new" class="text-primary hover:underline text-sm">← 새 백테스트</a>
			<h1 class="text-3xl font-bold mt-4">백테스트 결과</h1>
			<p class="text-muted-foreground mt-2">
				{result.symbol} {result.interval} {result.strategy}
			</p>
		</div>

		<!-- 성과 지표 카드 -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<Card class="p-6">
				<div class="text-sm text-muted-foreground mb-2">총 수익률</div>
				<div
					class="text-3xl font-bold {result.total_return >= 0
						? 'text-green-600'
						: 'text-red-600'}"
				>
					{formatPercent(result.total_return)}
				</div>
				<div class="text-sm text-muted-foreground mt-1">
					${formatNumber(result.final_equity - result.initial_balance)}
				</div>
			</Card>

			<Card class="p-6">
				<div class="text-sm text-muted-foreground mb-2">Sharpe Ratio</div>
				<div class="text-3xl font-bold">{result.sharpe_ratio.toFixed(2)}</div>
				<div class="text-sm text-muted-foreground mt-1">위험조정수익</div>
			</Card>

			<Card class="p-6">
				<div class="text-sm text-muted-foreground mb-2">Max Drawdown</div>
				<div class="text-3xl font-bold text-red-600">
					-{result.max_drawdown_pct.toFixed(2)}%
				</div>
				<div class="text-sm text-muted-foreground mt-1">최대 낙폭</div>
			</Card>

			<Card class="p-6">
				<div class="text-sm text-muted-foreground mb-2">승률</div>
				<div class="text-3xl font-bold">{result.win_rate.toFixed(2)}%</div>
				<div class="text-sm text-muted-foreground mt-1">
					{result.winning_trades}승 {result.losing_trades}패
				</div>
			</Card>
		</div>

		<!-- 자산 곡선 차트 -->
		{#if result.chart_data && result.chart_data.equity_curve.length > 0}
			<Card class="p-6 mb-8">
				<div bind:this={equityChartContainer} class="w-full bg-card" style="height: 450px;"></div>
			</Card>
		{/if}

		<!-- 가격 & 이평선 차트 (MA Cross 전략인 경우에만) -->
		{#if result.chart_data?.indicators}
			<Card class="p-6 mb-8">
				<div bind:this={priceChartContainer} class="w-full bg-card" style="height: 450px;"></div>
			</Card>
		{/if}

		<!-- 추가 지표 (접기/펼치기) -->
		<div class="mb-8">
			<Button variant="outline" onclick={() => (showAllMetrics = !showAllMetrics)}>
				{showAllMetrics ? '간단히 보기' : '더 많은 지표 보기'}
			</Button>

			{#if showAllMetrics}
				<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mt-4">
					<Card class="p-4">
						<div class="text-xs text-muted-foreground">총 거래 횟수</div>
						<div class="text-xl font-semibold mt-1">{result.total_trades}회</div>
					</Card>

					<Card class="p-4">
						<div class="text-xs text-muted-foreground">초기 자금</div>
						<div class="text-xl font-semibold mt-1">${formatNumber(result.initial_balance)}</div>
					</Card>

					<Card class="p-4">
						<div class="text-xs text-muted-foreground">최종 자금</div>
						<div class="text-xl font-semibold mt-1">${formatNumber(result.final_equity)}</div>
					</Card>

					<Card class="p-4">
						<div class="text-xs text-muted-foreground">수수료율</div>
						<div class="text-xl font-semibold mt-1">{(result.commission * 100).toFixed(2)}%</div>
					</Card>

					<Card class="p-4">
						<div class="text-xs text-muted-foreground">사용된 캔들</div>
						<div class="text-xl font-semibold mt-1">{result.candles_used}개</div>
					</Card>

					<Card class="p-4">
						<div class="text-xs text-muted-foreground">기간</div>
						<div class="text-xl font-semibold mt-1">{result.duration}</div>
					</Card>
				</div>
			{/if}
		</div>

		<!-- 거래 내역 -->
		<Card class="p-6">
			<h2 class="text-xl font-semibold mb-4">
				거래 내역
				<span class="text-sm text-muted-foreground font-normal">(전체 {result.total_trades}개)</span>
			</h2>

			{#if result.recent_trades && result.recent_trades.length > 0}
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class="border-b">
							<tr class="text-left">
								<th class="pb-3 font-semibold text-sm">날짜 & 시간</th>
								<th class="pb-3 font-semibold text-sm">구분</th>
								<th class="pb-3 font-semibold text-sm text-right">가격</th>
								<th class="pb-3 font-semibold text-sm text-right">수량</th>
								<th class="pb-3 font-semibold text-sm text-right">수수료</th>
								<th class="pb-3 font-semibold text-sm text-right">잔액</th>
								<th class="pb-3 font-semibold text-sm">이유</th>
							</tr>
						</thead>
						<tbody>
							{#each result.recent_trades as trade}
								<tr class="border-b last:border-0 hover:bg-muted/20 transition-colors">
									<td class="py-3 text-sm">{formatDate(trade.timestamp)}</td>
									<td class="py-3">
										<span
											class="font-semibold text-sm {trade.side === 'BUY'
												? 'text-green-600'
												: 'text-red-600'}"
										>
											{trade.side}
										</span>
									</td>
									<td class="py-3 text-sm text-right">{formatNumber(trade.price)}</td>
									<td class="py-3 text-sm text-right">{trade.quantity.toFixed(4)}</td>
									<td class="py-3 text-sm text-right">${trade.fee.toFixed(2)}</td>
									<td class="py-3 text-sm text-right font-semibold">
										${formatNumber(trade.balance)}
									</td>
									<td class="py-3 text-xs text-muted-foreground max-w-xs truncate" title={trade.reason}>
										{trade.reason}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{:else}
				<div class="text-center py-8 text-muted-foreground">거래 내역이 없습니다</div>
			{/if}
		</Card>

		<!-- 액션 버튼 -->
		<div class="mt-8 flex gap-4 justify-center flex-wrap">
			<Button onclick={() => (window.location.href = '/backtest/new')}>
				다른 파라미터로 재실행
			</Button>
			<Button variant="outline" onclick={() => window.print()}>
				결과 출력
			</Button>
		</div>
	{/if}
</div>
