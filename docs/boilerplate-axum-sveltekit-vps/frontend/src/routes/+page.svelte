<script lang="ts">
	let { data } = $props();
</script>

<svelte:head>
	<title>App — SSR</title>
</svelte:head>

<h1>Axum + SvelteKit</h1>
<p class="muted">Página renderizada no servidor (SSR) — boa para SEO.</p>

<div class="card">
	<p><strong>API health:</strong> {data.healthOk ? 'ok' : 'offline'}</p>
	{#if data.hello}
		<p><strong>Mensagem:</strong> {data.hello.message}</p>
		<p class="muted">{data.hello.ts}</p>
	{:else}
		<p class="muted">Não foi possível carregar /api/hello no SSR.</p>
	{/if}
</div>

<p>
	<a href="/dashboard">Abrir dashboard real-time →</a>
</p>
