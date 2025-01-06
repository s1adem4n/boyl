<script lang="ts">
	import remote from '$lib/remote';
	import { clientState } from '$lib/state.svelte';
	import type { ClientGame, Game } from '$lib/types';
	import VanillaTilt from 'vanilla-tilt';
	import { Tilted } from './ui';

	interface GameCard {
		game: Game;
		clientGame: ClientGame;
	}

	let { game, clientGame }: GameCard = $props();
</script>

<div class="flex flex-col gap-2">
	<Tilted>
		<button
			class="before:border-1 animate-fade-in-down group relative h-full overflow-hidden rounded before:absolute before:inset-0 before:z-10 before:rounded before:border-white/20"
			onclick={() => clientState.launchGame(clientGame.id)}
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
		</button>
	</Tilted>
	<div class="flex flex-col">
		<span class="truncate text-lg font-bold">{game.name}</span>
		<span class="text-muted">
			{new Date(game.released).getFullYear()}
		</span>
	</div>
</div>
