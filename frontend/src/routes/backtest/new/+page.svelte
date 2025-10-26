<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Label from '$lib/components/ui/Label.svelte';
	import Select from '$lib/components/ui/Select.svelte';

	// Form state
	let symbol = $state('BTCUSDT');
	let interval = $state('1h');
	let startDate = $state('2025-07-01');
	let endDate = $state('2025-10-17');
	let initialBalance = $state(10000);
	let commission = $state(0.001);
	let positionSize = $state(0.01);

	// Strategy selection
	type Strategy = 'ma_cross' | 'rsi' | 'bb_rsi' | 'dca' | 'golden_rsi_bb';
	let selectedStrategy = $state<Strategy>('golden_rsi_bb');

	// Strategy parameters
	let fastPeriod = $state(10);
	let slowPeriod = $state(30);
	let rsiPeriod = $state(14);
	let rsiOversold = $state(30);
	let rsiOverbought = $state(70);
	let bbPeriod = $state(20);
	let bbStdDev = $state(2);
	let dcaPeriod = $state('24h');
	let dcaAmountUSDT = $state(100);
	
	// Golden RSI BB strategy parameters
	let goldenFastPeriod = $state(5);
	let goldenSlowPeriod = $state(20);
	let goldenRsiPeriod = $state(14);
	let goldenRsiLowerBound = $state(40);
	let goldenRsiUpperBound = $state(70);
	let goldenBbPeriod = $state(20);
	let goldenBbMultiplier = $state(2.0);
	let goldenVolumeThreshold = $state(1.3);
	let goldenTakeProfitPct = $state(0.06);
	let goldenStopLossPct = $state(0.03);

	let isLoading = $state(false);
	let error = $state('');
	let dataValidation = $state<{
		isChecking: boolean;
		hasData: boolean;
		isComplete: boolean;
		message: string;
	} | null>(null);

	const symbolOptions = [
		{ value: 'BTCUSDT', label: 'BTC/USDT' },
		{ value: 'ETHUSDT', label: 'ETH/USDT' },
		{ value: 'BNBUSDT', label: 'BNB/USDT' }
	];

	const intervalOptions = [
		{ value: '1m', label: '1분' },
		{ value: '5m', label: '5분' },
		{ value: '15m', label: '15분' },
		{ value: '30m', label: '30분' },
		{ value: '1h', label: '1시간' },
		{ value: '4h', label: '4시간' },
		{ value: '1d', label: '1일' }
	];

	// Validate data availability
	async function validateData() {
		if (!symbol || !interval || !startDate || !endDate) {
			dataValidation = null;
			return;
		}

		dataValidation = {
			isChecking: true,
			hasData: false,
			isComplete: false,
			message: 'Checking data availability...'
		};

		try {
			const response = await fetch(
				`http://localhost:8080/api/v1/data/validate?symbol=${symbol}&interval=${interval}&start=${startDate}&end=${endDate}`
			);

			if (!response.ok) {
				throw new Error('Failed to validate data');
			}

			const result = await response.json();
			const data = result.data || result;

			dataValidation = {
				isChecking: false,
				hasData: data.has_data,
				isComplete: data.is_complete,
				message: data.message
			};
		} catch (err) {
			console.error('Data validation error:', err);
			dataValidation = {
				isChecking: false,
				hasData: false,
				isComplete: false,
				message: 'Failed to check data availability'
			};
		}
	}

	// Watch for changes and validate data
	$effect(() => {
		// Trigger validation when key parameters change
		const _ = symbol + interval + startDate + endDate;
		validateData();
	});

	async function runBacktest() {
		isLoading = true;
		error = '';

		try {
			// Build request body based on selected strategy
			const requestBody: any = {
				symbol,
				interval,
				start_date: startDate,
				end_date: endDate,
				initial_balance: initialBalance,
				commission,
				strategy: selectedStrategy,
				position_size: positionSize
			};

			// Add strategy-specific parameters
			if (selectedStrategy === 'ma_cross') {
				requestBody.fast_period = fastPeriod;
				requestBody.slow_period = slowPeriod;
			} else if (selectedStrategy === 'rsi') {
				requestBody.rsi_period = rsiPeriod;
				requestBody.rsi_oversold = rsiOversold;
				requestBody.rsi_overbought = rsiOverbought;
			} else if (selectedStrategy === 'bb_rsi') {
				requestBody.bb_period = bbPeriod;
				requestBody.bb_std_dev = bbStdDev;
				requestBody.rsi_period = rsiPeriod;
				requestBody.rsi_oversold = rsiOversold;
				requestBody.rsi_overbought = rsiOverbought;
			} else if (selectedStrategy === 'dca') {
				requestBody.dca_period = dcaPeriod;
				requestBody.dca_amount_usdt = dcaAmountUSDT;
			} else if (selectedStrategy === 'golden_rsi_bb') {
				requestBody.golden_fast_period = goldenFastPeriod;
				requestBody.golden_slow_period = goldenSlowPeriod;
				requestBody.golden_rsi_period = goldenRsiPeriod;
				requestBody.golden_rsi_lower_bound = goldenRsiLowerBound;
				requestBody.golden_rsi_upper_bound = goldenRsiUpperBound;
				requestBody.golden_bb_period = goldenBbPeriod;
				requestBody.golden_bb_multiplier = goldenBbMultiplier;
				requestBody.golden_volume_threshold = goldenVolumeThreshold;
				requestBody.golden_take_profit_pct = goldenTakeProfitPct;
				requestBody.golden_stop_loss_pct = goldenStopLossPct;
			}

			console.log('🚀 백테스트 요청:', requestBody);

			const response = await fetch('http://localhost:8080/api/v1/backtest/run', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(requestBody)
			});

			console.log('📡 응답 상태:', response.status, response.statusText);

			if (!response.ok) {
				const errorText = await response.text();
				console.error('❌ 에러 응답:', errorText);
				throw new Error(`백테스트 실행 실패 (${response.status}): ${errorText}`);
			}

			const result = await response.json();
			console.log('✅ 백테스트 결과:', result);
			
			// Extract data from wrapper
			const backtestData = result.data || result;
			console.log('📊 실제 데이터:', backtestData);
			
			// Store result in sessionStorage and navigate
			sessionStorage.setItem('backtest_result', JSON.stringify(backtestData));
			console.log('💾 sessionStorage에 저장 완료');
			
			window.location.href = '/backtest/result';
		} catch (err) {
			console.error('🔥 백테스트 실행 에러:', err);
			error = err instanceof Error ? err.message : '알 수 없는 오류가 발생했습니다';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="container mx-auto p-4 md:p-8 max-w-7xl">
	<div class="mb-8">
		<a href="/" class="text-primary hover:underline text-sm">← 홈으로</a>
		<h1 class="text-3xl font-bold mt-4">백테스트 실행</h1>
		<p class="text-muted-foreground mt-2">
			트레이딩 전략을 테스트하고 과거 데이터로 성과를 분석하세요
		</p>
	</div>

	{#if error}
		<div class="mb-6 p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
			<p class="text-destructive font-semibold">{error}</p>
		</div>
	{/if}

	{#if dataValidation}
		<div
			class="mb-6 p-4 rounded-lg border {dataValidation.isChecking
				? 'bg-muted/50 border-border'
				: dataValidation.isComplete
					? 'bg-green-50 border-green-200 dark:bg-green-950/50 dark:border-green-800'
					: dataValidation.hasData
						? 'bg-yellow-50 border-yellow-200 dark:bg-yellow-950/50 dark:border-yellow-800'
						: 'bg-red-50 border-red-200 dark:bg-red-950/50 dark:border-red-800'}"
		>
			<div class="flex items-start gap-3">
				<div class="text-2xl">
					{#if dataValidation.isChecking}
						⏳
					{:else if dataValidation.isComplete}
						✅
					{:else if dataValidation.hasData}
						⚠️
					{:else}
						❌
					{/if}
				</div>
				<div class="flex-1">
					<p class="font-semibold mb-1">
						{#if dataValidation.isChecking}
							데이터 확인 중...
						{:else if dataValidation.isComplete}
							데이터 준비 완료
						{:else if dataValidation.hasData}
							데이터 부족
						{:else}
							데이터 없음
						{/if}
					</p>
					<p class="text-sm opacity-90">{dataValidation.message}</p>
					{#if !dataValidation.isChecking && !dataValidation.isComplete}
						<p class="text-sm mt-2 font-semibold">
							💡 해결 방법: 백엔드에서 데이터를 수집해주세요
						</p>
						<code class="text-xs bg-black/10 dark:bg-white/10 px-2 py-1 rounded mt-1 inline-block">
							cd backend && ./bin/collector -symbol {symbol} -interval {interval} -days 120
						</code>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		<!-- 왼쪽: 기본 설정 & 전략 선택 -->
		<div class="space-y-6">
			<!-- 기본 설정 -->
			<Card class="p-6">
				<h2 class="text-xl font-semibold mb-4">기본 설정</h2>

				<div class="space-y-4">
					<div>
						<Label for="symbol">거래 쌍</Label>
						<Select id="symbol" bind:value={symbol} options={symbolOptions} class="mt-1.5" />
					</div>

					<div>
						<Label for="interval">시간 간격</Label>
						<Select id="interval" bind:value={interval} options={intervalOptions} class="mt-1.5" />
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<Label for="startDate">시작일</Label>
							<Input id="startDate" type="date" bind:value={startDate} class="mt-1.5" />
						</div>
						<div>
							<Label for="endDate">종료일</Label>
							<Input id="endDate" type="date" bind:value={endDate} class="mt-1.5" />
						</div>
					</div>

					<div>
						<Label for="initialBalance">초기 자금 (USDT)</Label>
						<Input
							id="initialBalance"
							type="number"
							bind:value={initialBalance}
							min="0"
							step="1000"
							class="mt-1.5"
						/>
					</div>

					<div>
						<Label for="commission">거래 수수료 (%)</Label>
						<Input
							id="commission"
							type="number"
							bind:value={commission}
							min="0"
							max="1"
							step="0.0001"
							class="mt-1.5"
						/>
						<p class="text-xs text-muted-foreground mt-1">
							현재: {(commission * 100).toFixed(2)}%
						</p>
					</div>

					<div>
						<Label for="positionSize">포지션 크기 (비율)</Label>
						<Input
							id="positionSize"
							type="number"
							bind:value={positionSize}
							min="0.001"
							max="1"
							step="0.001"
							class="mt-1.5"
						/>
						<p class="text-xs text-muted-foreground mt-1">
							전체 자금의 {(positionSize * 100).toFixed(1)}%
						</p>
					</div>
				</div>
			</Card>

			<!-- 전략 선택 -->
			<Card class="p-6">
				<h2 class="text-xl font-semibold mb-4">전략 선택</h2>

				<div class="space-y-3">
					<button
						class="w-full text-left p-4 rounded-lg border-2 transition-all {selectedStrategy ===
						'ma_cross'
							? 'border-primary bg-primary/5'
							: 'border-border hover:border-primary/50'}"
						onclick={() => (selectedStrategy = 'ma_cross')}
					>
						<div class="font-semibold">이동평균 교차 (MA Cross)</div>
						<p class="text-sm text-muted-foreground mt-1">
							빠른/느린 이동평균선의 교차로 매매 신호 생성
						</p>
					</button>

					<button
						class="w-full text-left p-4 rounded-lg border-2 transition-all {selectedStrategy ===
						'rsi'
							? 'border-primary bg-primary/5'
							: 'border-border hover:border-primary/50'}"
						onclick={() => (selectedStrategy = 'rsi')}
					>
						<div class="font-semibold">RSI 전략</div>
						<p class="text-sm text-muted-foreground mt-1">
							상대강도지수로 과매수/과매도 구간 판단
						</p>
					</button>

					<button
						class="w-full text-left p-4 rounded-lg border-2 transition-all {selectedStrategy ===
						'bb_rsi'
							? 'border-primary bg-primary/5'
							: 'border-border hover:border-primary/50'}"
						onclick={() => (selectedStrategy = 'bb_rsi')}
					>
						<div class="font-semibold">볼린저밴드 + RSI</div>
						<p class="text-sm text-muted-foreground mt-1">
							볼린저밴드와 RSI를 조합한 복합 전략
						</p>
					</button>

					<button
						class="w-full text-left p-4 rounded-lg border-2 transition-all {selectedStrategy ===
						'dca'
							? 'border-primary bg-primary/5'
							: 'border-border hover:border-primary/50'}"
						onclick={() => (selectedStrategy = 'dca')}
					>
						<div class="font-semibold">적립식 투자 (DCA)</div>
						<p class="text-sm text-muted-foreground mt-1">
							일정 기간마다 고정 금액을 자동 매수
						</p>
					</button>

					<button
						class="w-full text-left p-4 rounded-lg border-2 transition-all {selectedStrategy ===
						'golden_rsi_bb'
							? 'border-primary bg-primary/5'
							: 'border-border hover:border-primary/50'}"
						onclick={() => (selectedStrategy = 'golden_rsi_bb')}
					>
						<div class="font-semibold">🎯 골든크로스 + RSI + 볼린저밴드</div>
						<p class="text-sm text-muted-foreground mt-1">
							다중 지표 확인 + 거래량 필터 + 명확한 익절/손절 (고급)
						</p>
					</button>
				</div>
			</Card>
		</div>

		<!-- 오른쪽: 전략 파라미터 -->
		<div>
			<Card class="p-6">
				<h2 class="text-xl font-semibold mb-4">전략 파라미터</h2>

				{#if selectedStrategy === 'ma_cross'}
					<div class="space-y-4">
						<div>
							<Label for="fastPeriod">빠른 이동평균 기간</Label>
							<Input
								id="fastPeriod"
								type="number"
								bind:value={fastPeriod}
								min="1"
								max="100"
								class="mt-1.5"
							/>
							<p class="text-xs text-muted-foreground mt-1">단기 추세를 따르는 이동평균선</p>
						</div>

						<div>
							<Label for="slowPeriod">느린 이동평균 기간</Label>
							<Input
								id="slowPeriod"
								type="number"
								bind:value={slowPeriod}
								min="1"
								max="200"
								class="mt-1.5"
							/>
							<p class="text-xs text-muted-foreground mt-1">장기 추세를 따르는 이동평균선</p>
						</div>

						<div class="bg-muted/50 p-4 rounded-lg">
							<p class="text-sm">
								<strong>매수 신호:</strong> 빠른 MA가 느린 MA를 상향 돌파<br />
								<strong>매도 신호:</strong> 빠른 MA가 느린 MA를 하향 돌파
							</p>
						</div>
					</div>
				{:else if selectedStrategy === 'rsi'}
					<div class="space-y-4">
						<div>
							<Label for="rsiPeriod">RSI 기간</Label>
							<Input
								id="rsiPeriod"
								type="number"
								bind:value={rsiPeriod}
								min="2"
								max="50"
								class="mt-1.5"
							/>
						</div>

						<div>
							<Label for="rsiOversold">과매도 레벨</Label>
							<Input
								id="rsiOversold"
								type="number"
								bind:value={rsiOversold}
								min="0"
								max="50"
								class="mt-1.5"
							/>
							<p class="text-xs text-muted-foreground mt-1">이 값 이하면 매수 신호</p>
						</div>

						<div>
							<Label for="rsiOverbought">과매수 레벨</Label>
							<Input
								id="rsiOverbought"
								type="number"
								bind:value={rsiOverbought}
								min="50"
								max="100"
								class="mt-1.5"
							/>
							<p class="text-xs text-muted-foreground mt-1">이 값 이상이면 매도 신호</p>
						</div>

						<div class="bg-muted/50 p-4 rounded-lg">
							<p class="text-sm">
								<strong>매수 신호:</strong> RSI가 {rsiOversold} 이하에서 상승<br />
								<strong>매도 신호:</strong> RSI가 {rsiOverbought} 이상에서 하락
							</p>
						</div>
					</div>
				{:else if selectedStrategy === 'bb_rsi'}
					<div class="space-y-4">
						<div>
							<Label for="bbPeriod">볼린저밴드 기간</Label>
							<Input
								id="bbPeriod"
								type="number"
								bind:value={bbPeriod}
								min="2"
								max="100"
								class="mt-1.5"
							/>
						</div>

						<div>
							<Label for="bbStdDev">표준편차 배수</Label>
							<Input
								id="bbStdDev"
								type="number"
								bind:value={bbStdDev}
								min="0.5"
								max="5"
								step="0.1"
								class="mt-1.5"
							/>
						</div>

						<div>
							<Label for="rsiPeriodBB">RSI 기간</Label>
							<Input
								id="rsiPeriodBB"
								type="number"
								bind:value={rsiPeriod}
								min="2"
								max="50"
								class="mt-1.5"
							/>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div>
								<Label for="rsiOversoldBB">RSI 과매도</Label>
								<Input
									id="rsiOversoldBB"
									type="number"
									bind:value={rsiOversold}
									min="0"
									max="50"
									class="mt-1.5"
								/>
							</div>
							<div>
								<Label for="rsiOverboughtBB">RSI 과매수</Label>
								<Input
									id="rsiOverboughtBB"
									type="number"
									bind:value={rsiOverbought}
									min="50"
									max="100"
									class="mt-1.5"
								/>
							</div>
						</div>

						<div class="bg-muted/50 p-4 rounded-lg">
							<p class="text-sm">
								가격이 볼린저밴드 하단에 접근하고 RSI가 과매도 상태이면 매수,<br />
								상단에 접근하고 RSI가 과매수 상태이면 매도
							</p>
						</div>
					</div>
				{:else if selectedStrategy === 'dca'}
					<div class="space-y-4">
						<div>
							<Label for="dcaPeriod">구매 주기</Label>
							<select
								id="dcaPeriod"
								bind:value={dcaPeriod}
								class="w-full px-3 py-2 border border-input rounded-md bg-background mt-1.5"
							>
								<option value="1h">1시간마다</option>
								<option value="4h">4시간마다</option>
								<option value="12h">12시간마다</option>
								<option value="24h">1일마다</option>
								<option value="168h">7일마다 (주간)</option>
								<option value="720h">30일마다 (월간)</option>
							</select>
							<p class="text-xs text-muted-foreground mt-1">매수를 실행할 시간 간격</p>
						</div>

						<div>
							<Label for="dcaAmountUSDT">구매 금액 (USDT)</Label>
							<Input
								id="dcaAmountUSDT"
								type="number"
								bind:value={dcaAmountUSDT}
								min="1"
								step="10"
								class="mt-1.5"
							/>
							<p class="text-xs text-muted-foreground mt-1">
								매번 구매할 고정 금액 (USDT 기준)
							</p>
						</div>

						<div class="bg-muted/50 p-4 rounded-lg">
							<p class="text-sm">
								<strong>적립식 투자 (DCA)</strong><br />
								시장 상황과 무관하게 {dcaPeriod === '1h' ? '1시간' : dcaPeriod === '4h' ? '4시간' : dcaPeriod === '12h' ? '12시간' : dcaPeriod === '24h' ? '매일' : dcaPeriod === '168h' ? '매주' : '매달'}마다 {dcaAmountUSDT} USDT를 자동으로 매수합니다.<br />
								<small class="text-muted-foreground">
									※ 가격 변동성을 분산시켜 평균 매수가를 낮추는 전략
								</small>
							</p>
						</div>
					</div>
				{:else if selectedStrategy === 'golden_rsi_bb'}
					<div class="space-y-4">
						<div class="bg-primary/10 p-4 rounded-lg border border-primary/20 mb-4">
							<p class="text-sm font-semibold mb-2">🎯 고급 복합 전략</p>
							<p class="text-xs text-muted-foreground">
								4가지 진입 조건 + 3가지 청산 조건을 사용하는 엄격한 전략입니다.
							</p>
						</div>

						<div class="border-t pt-4">
							<h3 class="font-semibold mb-3 text-sm">📈 이동평균 (골든/데드 크로스)</h3>
							<div class="grid grid-cols-2 gap-4">
								<div>
									<Label for="goldenFastPeriod">빠른 MA 기간</Label>
									<Input
										id="goldenFastPeriod"
										type="number"
										bind:value={goldenFastPeriod}
										min="1"
										max="50"
										class="mt-1.5"
									/>
									<p class="text-xs text-muted-foreground mt-1">기본: 5일선</p>
								</div>
								<div>
									<Label for="goldenSlowPeriod">느린 MA 기간</Label>
									<Input
										id="goldenSlowPeriod"
										type="number"
										bind:value={goldenSlowPeriod}
										min="1"
										max="200"
										class="mt-1.5"
									/>
									<p class="text-xs text-muted-foreground mt-1">기본: 20일선</p>
								</div>
							</div>
						</div>

						<div class="border-t pt-4">
							<h3 class="font-semibold mb-3 text-sm">📊 RSI 필터</h3>
							<div>
								<Label for="goldenRsiPeriod">RSI 기간</Label>
								<Input
									id="goldenRsiPeriod"
									type="number"
									bind:value={goldenRsiPeriod}
									min="2"
									max="50"
									class="mt-1.5"
								/>
							</div>
							<div class="grid grid-cols-2 gap-4 mt-4">
								<div>
									<Label for="goldenRsiLowerBound">RSI 하한선</Label>
									<Input
										id="goldenRsiLowerBound"
										type="number"
										bind:value={goldenRsiLowerBound}
										min="0"
										max="100"
										class="mt-1.5"
									/>
									<p class="text-xs text-muted-foreground mt-1">이 값 이상이어야 진입</p>
								</div>
								<div>
									<Label for="goldenRsiUpperBound">RSI 상한선</Label>
									<Input
										id="goldenRsiUpperBound"
										type="number"
										bind:value={goldenRsiUpperBound}
										min="0"
										max="100"
										class="mt-1.5"
									/>
									<p class="text-xs text-muted-foreground mt-1">이 값 이하여야 진입</p>
								</div>
							</div>
							<p class="text-xs text-muted-foreground mt-2">
								💡 RSI {goldenRsiLowerBound}-{goldenRsiUpperBound} 구간에서만 진입
							</p>
						</div>

						<div class="border-t pt-4">
							<h3 class="font-semibold mb-3 text-sm">📉 볼린저밴드</h3>
							<div class="grid grid-cols-2 gap-4">
								<div>
									<Label for="goldenBbPeriod">BB 기간</Label>
									<Input
										id="goldenBbPeriod"
										type="number"
										bind:value={goldenBbPeriod}
										min="2"
										max="100"
										class="mt-1.5"
									/>
								</div>
								<div>
									<Label for="goldenBbMultiplier">표준편차 배수</Label>
									<Input
										id="goldenBbMultiplier"
										type="number"
										bind:value={goldenBbMultiplier}
										min="0.5"
										max="5"
										step="0.1"
										class="mt-1.5"
									/>
								</div>
							</div>
							<p class="text-xs text-muted-foreground mt-2">
								💡 가격이 BB 중간선 위에 있어야 진입
							</p>
						</div>

						<div class="border-t pt-4">
							<h3 class="font-semibold mb-3 text-sm">📦 거래량 필터</h3>
							<div>
								<Label for="goldenVolumeThreshold">거래량 배율</Label>
								<Input
									id="goldenVolumeThreshold"
									type="number"
									bind:value={goldenVolumeThreshold}
									min="1.0"
									max="3.0"
									step="0.1"
									class="mt-1.5"
								/>
								<p class="text-xs text-muted-foreground mt-1">
									평균 거래량의 {goldenVolumeThreshold}배 이상이어야 진입
								</p>
							</div>
						</div>

						<div class="border-t pt-4">
							<h3 class="font-semibold mb-3 text-sm">💰 익절/손절</h3>
							<div class="grid grid-cols-2 gap-4">
								<div>
									<Label for="goldenTakeProfitPct">익절 (%)</Label>
									<Input
										id="goldenTakeProfitPct"
										type="number"
										bind:value={goldenTakeProfitPct}
										min="0.01"
										max="0.50"
										step="0.01"
										class="mt-1.5"
									/>
									<p class="text-xs text-muted-foreground mt-1">
										+{(goldenTakeProfitPct * 100).toFixed(1)}% 도달 시 매도
									</p>
								</div>
								<div>
									<Label for="goldenStopLossPct">손절 (%)</Label>
									<Input
										id="goldenStopLossPct"
										type="number"
										bind:value={goldenStopLossPct}
										min="0.01"
										max="0.20"
										step="0.01"
										class="mt-1.5"
									/>
									<p class="text-xs text-muted-foreground mt-1">
										-{(goldenStopLossPct * 100).toFixed(1)}% 도달 시 손절
									</p>
								</div>
							</div>
							<p class="text-xs text-muted-foreground mt-2">
								💡 리스크-리워드 비율: 1:{(goldenTakeProfitPct / goldenStopLossPct).toFixed(1)}
							</p>
						</div>

						<div class="bg-muted/50 p-4 rounded-lg">
							<p class="text-sm">
								<strong>진입 조건 (모두 만족 필요):</strong><br />
								✅ MA{goldenFastPeriod} &gt; MA{goldenSlowPeriod} (골든크로스)<br />
								✅ RSI {goldenRsiLowerBound}-{goldenRsiUpperBound} 구간<br />
								✅ 가격 &gt; 볼린저 중간선<br />
								✅ 거래량 &gt;= 평균 × {goldenVolumeThreshold}<br /><br />
								<strong>청산 조건 (하나만 만족):</strong><br />
								💰 익절: +{(goldenTakeProfitPct * 100).toFixed(0)}%<br />
								🛑 손절: -{(goldenStopLossPct * 100).toFixed(0)}%<br />
								📉 데드크로스: MA{goldenFastPeriod} &lt; MA{goldenSlowPeriod}
							</p>
						</div>
					</div>
				{/if}
			</Card>
		</div>
	</div>

	<!-- 실행 버튼 -->
	<div class="mt-8 flex justify-center flex-col items-center gap-2">
		<Button
			size="lg"
			onclick={runBacktest}
			disabled={isLoading || (dataValidation !== null && !dataValidation.isComplete)}
			class="min-w-[200px]"
		>
			{#if isLoading}
				<span class="animate-spin mr-2">⏳</span>
				실행 중...
			{:else}
				🚀 백테스트 실행
			{/if}
		</Button>
		{#if dataValidation && !dataValidation.isComplete && !dataValidation.isChecking}
			<p class="text-sm text-muted-foreground">
				⚠️ 데이터가 부족하여 백테스트를 실행할 수 없습니다
			</p>
		{/if}
	</div>
</div>
