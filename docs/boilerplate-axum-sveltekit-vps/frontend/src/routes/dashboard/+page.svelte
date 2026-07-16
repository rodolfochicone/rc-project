<script lang="ts">
	import { onMount } from 'svelte';
	import { wsUrl } from '$lib/config';

	type Sample = {
		name: string;
		value: number;
		recorded_at: string;
	};

	let connected = $state(false);
	let latest = $state<Sample | null>(null);
	let history = $state<Sample[]>([]);
	let error = $state<string | null>(null);

	onMount(() => {
		const url = wsUrl();
		const ws = new WebSocket(url);

		ws.onopen = () => {
			connected = true;
			error = null;
		};
		ws.onclose = () => {
			connected = false;
		};
		ws.onerror = () => {
			error = `Falha no WebSocket: ${url}`;
		};
		ws.onmessage = (ev) => {
			try {
				const sample = JSON.parse(ev.data) as Sample;
				latest = sample;
				history = [sample, ...history].slice(0, 20);
			} catch {
				// ignore
			}
		};

		return () => ws.close();
	});
</script>

<svelte:head>
	<title>Dashboard ao vivo</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<h1>Dashboard ao vivo</h1>
<p class="muted">
	WebSocket → Axum <code>/ws/metrics</code>
	{#if connected}
		· <span style="color:#5ddea5">conectado</span>
	{:else}
		· desconectado
	{/if}
</p>

{#if error}
	<div class="card"><p style="color:#ff8a8a">{error}</p></div>
{/if}

<div class="card">
	<p class="muted">{latest?.name ?? 'cpu_mock'}</p>
	<p class="metric">{latest ? latest.value.toFixed(1) : '—'}</p>
	<p class="muted">{latest?.recorded_at ?? 'aguardando samples…'}</p>
</div>

<div class="card">
	<p><strong>Últimos samples</strong></p>
	<ul>
		{#each history as row}
			<li class="muted">
				{row.value.toFixed(1)} · {row.recorded_at}
			</li>
		{/each}
	</ul>
</div>
