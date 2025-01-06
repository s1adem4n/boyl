<script lang="ts">
	import remote from '$lib/remote';
	import { clientState, remoteState } from '$lib/state.svelte';
</script>

<div class="grid auto-rows-fr grid-cols-[repeat(auto-fit,minmax(200px,250px))] gap-4 p-4">
	{#each clientState.games as clientGame}
		{@const game = remoteState.games.find((g) => g.id === clientGame.game)!}
		<div class="flex flex-col gap-1">
			<div
				class="before:border-1 animate-fade-in-down group relative h-full overflow-hidden rounded before:absolute before:inset-0 before:z-10 before:rounded before:border-white/20"
			>
				<img
					src={remote.files.getURL(game, game.cover)}
					class={[
						'h-full w-full',
						clientState.settings.fitCovers ? 'object-contain' : 'object-cover'
					]}
					loading="lazy"
					alt={game.name}
				/>
				<img
					src={remote.files.getURL(game, game.cover)}
					class="absolute inset-0 -z-10 h-full w-full object-cover blur-md"
					alt={game.name}
				/>
			</div>
			<span class="truncate">{game.name}</span>
		</div>
	{/each}
</div>
