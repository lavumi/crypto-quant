<script lang="ts">
	import { onMount } from 'svelte';

	let apiStatus = $state('Checking...');
	let isDarkMode = $state(false);
	
	// Check backend API status
	async function checkAPI() {
		try {
			const response = await fetch('http://localhost:8080/health');
			if (response.ok) {
				apiStatus = '✅ Connected';
			} else {
				apiStatus = '❌ Error';
			}
		} catch {
			apiStatus = '⚠️ Offline';
		}
	}
	
	// Dark mode toggle
	function toggleDarkMode() {
		isDarkMode = !isDarkMode;
		if (typeof window !== 'undefined') {
			localStorage.setItem('darkMode', isDarkMode.toString());
			document.body.classList.toggle('dark', isDarkMode);
		}
	}

	// Load dark mode preference
	onMount(() => {
		const savedMode = localStorage.getItem('darkMode');
		if (savedMode === 'true') {
			isDarkMode = true;
			document.body.classList.add('dark');
		}
		checkAPI();
	});
</script>

<div class="container mx-auto p-8 max-w-7xl">
	<!-- Header with controls -->
	<div class="flex justify-end items-center gap-4 mb-4">
		<button
			class="w-12 h-12 rounded-full bg-card border-2 border-border flex items-center justify-center text-2xl cursor-pointer transition-all hover:scale-110 shadow-md"
			onclick={toggleDarkMode}
			title="Toggle dark mode"
		>
			{isDarkMode ? '☀️' : '🌙'}
		</button>
		{#if apiStatus === '✅ Connected'}
			<a
				href="http://localhost:8080/swagger/index.html"
				target="_blank"
				class="bg-primary text-primary-foreground px-6 py-3 rounded-lg font-semibold transition-all hover:bg-primary/90 hover:-translate-y-0.5 shadow-md"
				title="API Documentation"
			>
				📚 Swagger
			</a>
		{/if}
	</div>

	<header class="text-center mb-12">
		<h1 class="text-5xl font-bold mb-2">🚀 Crypto Quant Trading System</h1>
		<p class="text-xl text-muted-foreground">
			Full-stack quantitative trading platform for cryptocurrency markets
		</p>
	</header>

	<div class="bg-card rounded-xl p-6 mb-6 shadow-md border">
		<h2 class="text-xl font-semibold mb-3">Backend Status</h2>
		<p class="text-2xl font-semibold my-4">{apiStatus}</p>
		{#if apiStatus !== '✅ Connected'}
			<div class="bg-muted p-4 rounded-lg mt-4">
				<p>Start the backend API server: <code class="bg-muted-foreground/20 px-2 py-1 rounded text-sm">cd backend && ./bin/api</code></p>
			</div>
		{/if}
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-6 mb-6">
		<a
			href="/data"
			class="bg-card rounded-xl p-6 shadow-md border transition-all hover:-translate-y-1 hover:shadow-lg relative group"
		>
			<h3 class="text-xl font-semibold text-primary mb-2">📥 Data Collection</h3>
			<p class="text-muted-foreground">Collect historical data from Binance for backtesting</p>
			<span
				class="absolute bottom-4 right-4 text-2xl text-primary opacity-0 group-hover:opacity-100 group-hover:translate-x-1 transition-all"
				>→</span
			>
		</a>

		<a
			href="/backtest/new"
			class="bg-card rounded-xl p-6 shadow-md border transition-all hover:-translate-y-1 hover:shadow-lg relative group"
		>
			<h3 class="text-xl font-semibold text-primary mb-2">📊 Backtesting</h3>
			<p class="text-muted-foreground">Test your trading strategies with historical data</p>
			<span
				class="absolute bottom-4 right-4 text-2xl text-primary opacity-0 group-hover:opacity-100 group-hover:translate-x-1 transition-all"
				>→</span
			>
		</a>

		<div class="bg-card rounded-xl p-6 shadow-md border opacity-60 cursor-not-allowed relative">
			<h3 class="text-xl font-semibold text-primary mb-2">📈 Live Data</h3>
			<p class="text-muted-foreground">Real-time cryptocurrency price monitoring</p>
			<span
				class="absolute top-4 right-4 text-xs bg-muted text-muted-foreground px-2 py-1 rounded"
				>Coming Soon</span
			>
		</div>

		<div class="bg-card rounded-xl p-6 shadow-md border opacity-60 cursor-not-allowed relative">
			<h3 class="text-xl font-semibold text-primary mb-2">💼 Portfolio</h3>
			<p class="text-muted-foreground">Track your virtual and live trading portfolios</p>
			<span
				class="absolute top-4 right-4 text-xs bg-muted text-muted-foreground px-2 py-1 rounded"
				>Coming Soon</span
			>
		</div>
	</div>

	<div class="bg-card rounded-xl p-6 shadow-md border">
		<h2 class="text-xl font-semibold mb-4">Quick Start</h2>
		<ol class="text-muted-foreground leading-8 space-y-2">
			<li>1. Start the backend API server</li>
			<li>2. Collect historical data with the collector</li>
			<li>3. Run your first backtest</li>
			<li>4. Analyze the results in this dashboard</li>
		</ol>
	</div>
</div>

