<script lang="ts">
	import { onMount } from 'svelte';

	type Trade = {
		timestamp: string;
		side: 'BUY' | 'SELL';
		price: number;
		quantity: number;
		fee: number;
	};

	type Result = {
		strategy: string;
		symbol: string;
		interval: string;
		total_return: number;
		sharpe_ratio: number;
		max_drawdown_pct: number;
		win_rate: number;
		total_trades: number;
		final_equity: number;
		recent_trades?: Trade[];
	};

	let result = $state<Result | null>(null);

	onMount(() => {
		const raw = sessionStorage.getItem('backtest_result');
		if (raw) result = JSON.parse(raw);
	});

	function money(value: number) {
		return `$${value.toLocaleString('en-US', { maximumFractionDigits: 2 })}`;
	}
</script>

{#if !result}
	<section class="rounded-lg border p-5">
		<h1 class="text-xl font-semibold">No result</h1>
		<a class="mt-3 inline-block text-sm text-primary hover:underline" href="/backtest/new">Run a backtest</a>
	</section>
{:else}
	<section class="space-y-5">
		<div class="flex flex-col gap-2 border-b pb-4 md:flex-row md:items-end md:justify-between">
			<div>
				<h1 class="text-2xl font-semibold">{result.strategy}</h1>
				<p class="mt-1 text-sm text-muted-foreground">{result.symbol} · {result.interval}</p>
			</div>
			<a class="text-sm text-primary hover:underline" href="/backtest/new">Run again</a>
		</div>

		<div class="grid gap-3 md:grid-cols-5">
			<div class="rounded-lg border p-4">
				<p class="text-xs uppercase text-muted-foreground">Return</p>
				<p class="mt-2 text-xl font-semibold">{result.total_return.toFixed(2)}%</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-xs uppercase text-muted-foreground">Sharpe</p>
				<p class="mt-2 text-xl font-semibold">{result.sharpe_ratio.toFixed(2)}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-xs uppercase text-muted-foreground">MDD</p>
				<p class="mt-2 text-xl font-semibold">{result.max_drawdown_pct.toFixed(2)}%</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-xs uppercase text-muted-foreground">Win Rate</p>
				<p class="mt-2 text-xl font-semibold">{result.win_rate.toFixed(2)}%</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-xs uppercase text-muted-foreground">Equity</p>
				<p class="mt-2 text-xl font-semibold">{money(result.final_equity)}</p>
			</div>
		</div>

		<div class="rounded-lg border">
			<div class="border-b px-4 py-3">
				<h2 class="font-medium">Trades</h2>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-left text-sm">
					<thead class="border-b text-muted-foreground">
						<tr>
							<th class="px-4 py-2 font-medium">Time</th>
							<th class="px-4 py-2 font-medium">Side</th>
							<th class="px-4 py-2 font-medium">Price</th>
							<th class="px-4 py-2 font-medium">Qty</th>
							<th class="px-4 py-2 font-medium">Fee</th>
						</tr>
					</thead>
					<tbody>
						{#each (result.recent_trades ?? []).slice(-25).reverse() as trade}
							<tr class="border-b last:border-0">
								<td class="px-4 py-2 text-muted-foreground">{new Date(trade.timestamp).toLocaleString()}</td>
								<td class="px-4 py-2 font-medium">{trade.side}</td>
								<td class="px-4 py-2">{money(trade.price)}</td>
								<td class="px-4 py-2">{trade.quantity.toFixed(6)}</td>
								<td class="px-4 py-2">{money(trade.fee)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</section>
{/if}
