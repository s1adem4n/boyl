<script lang="ts">
	import X from '~icons/lucide/x';

	import { Tween } from 'svelte/motion';
	import { clientState, remoteState } from '$lib/state.svelte';
	import { linear } from 'svelte/easing';
	import { formatBytes, sortCreated } from '$lib/utils';

	const progressColors = {
		starting: {
			bg: 'var(--color-gray-400)',
			fg: 'var(--color-gray-500)'
		},
		downloading: {
			bg: 'var(--color-blue-400)',
			fg: 'var(--color-blue-500)'
		},
		extracting: {
			bg: 'var(--color-purple-400)',
			fg: 'var(--color-purple-500)'
		},
		completed: {
			bg: 'var(--color-green-400)',
			fg: 'var(--color-green-500)'
		},
		failed: {
			bg: 'var(--color-red-400)',
			fg: 'var(--color-red-500)'
		}
	};
</script>

<div class="flex flex-col gap-4">
	<span class="text-lg">Downloads</span>
	{#each sortCreated(clientState.downloads) as download}
		{@const game = remoteState.games.find((g) => g.id === download.game)!}
		{@const progress = new Tween(download.progress, {
			duration: 1000,
			easing: linear
		})}
		<div class="flex flex-col gap-1">
			<span class="font-bold">{game.name}</span>
			<div
				class="flex h-16 w-full items-center gap-2 rounded p-4 text-sm text-white"
				style="background-image: linear-gradient(
          to right,
          {progressColors[download.status].fg} {progress.current * 100}%,
          {progressColors[download.status].bg} {progress.current * 100}%
        );"
			>
				<div class="flex flex-col">
					<span class="font-bold uppercase">{download.status}</span>
					{#if download.status === 'downloading' || download.status === 'extracting'}
						<span>
							{formatBytes(download.speed)}/s
						</span>
					{/if}
					{#if download.text}
						<span>{download.text}</span>
					{/if}
				</div>
				<span class="ml-auto">{formatBytes(download.total)}</span>
				<button onclick={() => clientState.cancelDownload(download.id)}>
					<X class="h-6 w-6" />
				</button>
			</div>
		</div>
	{:else}
		<p class="text-muted text-sm">No downloads. Try downloading a game from the browse page.</p>
	{/each}
</div>
