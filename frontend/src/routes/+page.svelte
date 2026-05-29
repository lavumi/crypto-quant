<script lang="ts">
	import { onMount } from 'svelte';

	let status = $state('checking');

	onMount(async () => {
		try {
			const response = await fetch('http://localhost:8080/health');
			status = response.ok ? 'online' : 'error';
		} catch {
			status = 'offline';
		}
	});
</script>

<section class="space-y-6">
	<div class="flex flex-col gap-3 border-b pb-5 md:flex-row md:items-end md:justify-between">
		<div>
			<h1 class="text-2xl font-semibold">Research Console</h1>
			<p class="mt-1 text-sm text-muted-foreground">백테스트와 실험 결과만 빠르게 확인하는 단순 화면입니다.</p>
		</div>
		<div class="text-sm">
			<span class="text-muted-foreground">API</span>
			<span class="ml-2 font-medium {status === 'online' ? 'text-emerald-600' : 'text-destructive'}">{status}</span>
		</div>
	</div>

	<div class="grid gap-3 md:grid-cols-2">
		<a class="rounded-lg border p-4 transition hover:bg-muted/60" href="/backtest/new">
			<h2 class="font-medium">Backtest</h2>
			<p class="mt-1 text-sm text-muted-foreground">전략, 기간, 자금만 입력해서 단일 백테스트를 실행합니다.</p>
		</a>
		<a class="rounded-lg border p-4 transition hover:bg-muted/60" href="/experiments">
			<h2 class="font-medium">Experiments</h2>
			<p class="mt-1 text-sm text-muted-foreground">저장된 single, sweep, walk-forward 결과를 목록으로 봅니다.</p>
		</a>
	</div>
</section>
