<script lang="ts">
	import Download from '~icons/lucide/download';

	import remote from '$lib/remote';
	import { clientState } from '$lib/state.svelte';
	import type { Game } from '$lib/types';
	import { Tilted } from './ui';

	interface GameCard {
		game: Game;
	}

	let { game }: GameCard = $props();
</script>

<div class="flex flex-col gap-1">
	<div
		class="before:border-1 animate-fade-in-down group relative h-full overflow-hidden rounded before:absolute before:inset-0 before:z-10 before:rounded before:border-white/20"
	>
		<img
			src={remote.files.getURL(game, game.cover)}
			class={['h-full w-full', clientState.settings.fitCovers ? 'object-contain' : 'object-cover']}
			alt={game.name}
		/>
		<img
			src={remote.files.getURL(game, game.cover)}
			class="absolute inset-0 -z-10 h-full w-full object-cover blur-md"
			alt={game.name}
		/>
	</div>
	<div class="flex w-full gap-2">
		<div class="flex flex-col overflow-hidden">
			<span class="truncate text-lg font-bold" title={game.name}>{game.name}</span>
			<span class="text-muted">
				{new Date(game.released).getFullYear()}
			</span>
		</div>
		<button class="ml-auto min-w-fit">
			<Download class="h-6 w-6" />
		</button>
	</div>
</div>
