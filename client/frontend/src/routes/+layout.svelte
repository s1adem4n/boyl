<script lang="ts">
	import { onMount } from 'svelte';
	import '../app.css';
	import { clientState } from '$lib/state.svelte';
	import { Spinner } from '$lib/components';
	import { goto } from '$app/navigation';
	import remote from '$lib/remote';

	let { children } = $props();
	let loading = $state(true);

	$effect(() => {
		if (loading) return;

		if (!clientState.settings.setup) {
			goto('/setup');
			return;
		}
		remote.baseURL = clientState.settings.serverUrl;
		if (clientState.settings.email && clientState.settings.password) {
			remote
				.collection('users')
				.authWithPassword(clientState.settings.email, clientState.settings.password)
				.then((res) => {
					console.log(res);
				})
				.catch((err) => {
					console.error(err);
				});
		}
	});

	onMount(() => {
		const unsub = clientState.load();

		unsub.then(() => (loading = false));

		return () => unsub.then((fn) => fn());
	});
</script>

{#if loading}
	<div class="absolute inset-0 flex items-center justify-center">
		<Spinner text="Loading client" />
	</div>
{:else}
	{@render children()}
{/if}
