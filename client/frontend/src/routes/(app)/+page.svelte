<script lang="ts">
	import { clientState, remoteState } from '$lib/state.svelte';
	import { InstalledGameCard } from '$lib/components';
</script>

<div class="grid auto-rows-fr grid-cols-[repeat(auto-fit,minmax(200px,250px))] gap-4 p-4">
	{#each clientState.games as clientGame}
		{@const game = remoteState.games.find((g) => g.id === clientGame.game)!}
		<InstalledGameCard {game} {clientGame} />
	{/each}
</div>
