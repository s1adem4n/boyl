<script>
	import Download from '~icons/lucide/download';
	import remote from '$lib/remote';
	import { clientState, remoteState } from '$lib/state.svelte';
</script>

<div class="grid auto-rows-fr grid-cols-[repeat(auto-fit,minmax(200px,1fr))] gap-4 p-4">
	{#each remoteState.games as game}
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
					alt={game.name}
				/>
				<img
					src={remote.files.getURL(game, game.cover)}
					class="absolute inset-0 -z-10 h-full w-full object-cover blur-md"
					alt={game.name}
				/>
			</div>
			<span class="truncate">{game.name}</span>
			<button
				class="flex items-center gap-1 rounded bg-gray-800 p-2 text-white"
				onclick={() => clientState.addDownload(game.id)}
			>
				<Download class="h-4 w-4" />
				<span>Download</span>
			</button>
		</div>
	{/each}
</div>
