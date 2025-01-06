<script>
	import { remoteState } from '$lib/state.svelte';
	import { BrowseGameCard } from '$lib/components';
</script>

<div class="grid auto-rows-fr grid-cols-[repeat(auto-fit,minmax(200px,1fr))] gap-4 p-4">
	{#each remoteState.games as game}
		<BrowseGameCard {game} />
	{/each}
</div>
