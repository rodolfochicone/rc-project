<script lang="ts">
	let { children } = $props();
</script>

<main>
	<nav>
		<a href="/">Home</a>
		<a href="/dashboard">Dashboard ao vivo</a>
	</nav>
	{@render children()}
</main>
