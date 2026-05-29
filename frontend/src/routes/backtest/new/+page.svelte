<script lang="ts">
	const symbols = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT'];
	const intervals = ['1h', '4h', '1d'];
	const strategies = ['golden_rsi_bb', 'ma_cross', 'rsi', 'bb_rsi', 'dca'];

	let symbol = $state('BTCUSDT');
	let interval = $state('1h');
	let strategy = $state('golden_rsi_bb');
	let startDate = $state('2024-01-01');
	let endDate = $state('2025-10-17');
	let initialBalance = $state(10000);
	let commission = $state(0.001);
	let positionSize = $state(1);
	let fastPeriod = $state(5);
	let slowPeriod = $state(20);
	let rsiPeriod = $state(14);
	let isLoading = $state(false);
	let error = $state('');

	async function runBacktest() {
		isLoading = true;
		error = '';

		const body = {
			symbol,
			interval,
			strategy,
			start_date: startDate,
			end_date: endDate,
			initial_balance: initialBalance,
			commission,
			position_size: positionSize,
			fast_period: fastPeriod,
			slow_period: slowPeriod,
			rsi_period: rsiPeriod,
			golden_fast_period: fastPeriod,
			golden_slow_period: slowPeriod,
			golden_rsi_period: rsiPeriod
		};

		try {
			const response = await fetch('http://localhost:8080/api/v1/backtest/run', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});
			const payload = await response.json();
			if (!response.ok) throw new Error(payload?.error ?? payload?.message ?? 'Backtest failed');

			sessionStorage.setItem('backtest_result', JSON.stringify(payload.data ?? payload));
			window.location.href = '/backtest/result';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			isLoading = false;
		}
	}
</script>

<section class="space-y-5">
	<div class="border-b pb-4">
		<h1 class="text-2xl font-semibold">Backtest</h1>
		<p class="mt-1 text-sm text-muted-foreground">필수 입력값만 두고 빠르게 실행합니다.</p>
	</div>

	{#if error}
		<div class="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">{error}</div>
	{/if}

	<form class="grid gap-4 md:grid-cols-4" onsubmit={(event) => { event.preventDefault(); runBacktest(); }}>
		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Symbol</span>
			<select class="h-10 w-full rounded-md border bg-background px-3" bind:value={symbol}>
				{#each symbols as item}<option value={item}>{item}</option>{/each}
			</select>
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Interval</span>
			<select class="h-10 w-full rounded-md border bg-background px-3" bind:value={interval}>
				{#each intervals as item}<option value={item}>{item}</option>{/each}
			</select>
		</label>

		<label class="space-y-1 text-sm md:col-span-2">
			<span class="text-muted-foreground">Strategy</span>
			<select class="h-10 w-full rounded-md border bg-background px-3" bind:value={strategy}>
				{#each strategies as item}<option value={item}>{item}</option>{/each}
			</select>
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Start</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="date" bind:value={startDate} />
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">End</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="date" bind:value={endDate} />
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Balance</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="number" bind:value={initialBalance} />
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Commission</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="number" step="0.0001" bind:value={commission} />
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Fast</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="number" bind:value={fastPeriod} />
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Slow</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="number" bind:value={slowPeriod} />
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">RSI</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="number" bind:value={rsiPeriod} />
		</label>

		<label class="space-y-1 text-sm">
			<span class="text-muted-foreground">Position</span>
			<input class="h-10 w-full rounded-md border bg-background px-3" type="number" step="0.01" bind:value={positionSize} />
		</label>

		<div class="md:col-span-4">
			<button class="h-10 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground disabled:opacity-50" disabled={isLoading}>
				{isLoading ? 'Running...' : 'Run backtest'}
			</button>
		</div>
	</form>
</section>
