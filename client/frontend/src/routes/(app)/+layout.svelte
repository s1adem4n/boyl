<script lang="ts">
	import Cog from '~icons/heroicons/cog';
	import LayoutGrid from '~icons/lucide/layout-grid';
	import Gamepad from '~icons/lucide/gamepad';
	import Download from '~icons/lucide/download';

	import { onMount } from 'svelte';
	import { remoteState } from '$lib/state.svelte';
	import { Downloads, Spinner } from '$lib/components';
	import { fade } from 'svelte/transition';
	import { page } from '$app/state';

	let { children } = $props();
	let loading = $state(true);
	let downloadsOpen = $state(true);

	onMount(() => {
		const unsub = remoteState.load();
		unsub.then(() => (loading = false));

		return () => unsub.then((fn) => fn());
	});

	function isActive(path: string) {
		return page.url.pathname === path && !downloadsOpen;
	}
</script>

{#if loading}
	<div out:fade class="absolute inset-0 flex h-screen items-center justify-center">
		<Spinner text="Connecting to remote server" />
	</div>
{:else}
	<div class="relative flex h-full w-full">
		<div
			class={[
				'bg-background border-border absolute inset-y-0 left-12 z-20 w-80 border-r px-4 py-3 transition-transform',
				downloadsOpen ? 'translate-x-0' : '-translate-x-full'
			]}
		>
			<Downloads />
		</div>
		<div class="border-border bg-background z-30 flex h-full flex-col border-r">
			<a href="/" class={['p-2 transition-colors', isActive('/') && 'bg-accent']}>
				<Gamepad class="h-8 w-8" />
			</a>
			<a href="/browse" class={['p-2 transition-colors', isActive('/browse') && 'bg-accent']}>
				<LayoutGrid class="h-8 w-8" />
			</a>
			<button
				class={['p-2 transition-colors', downloadsOpen && 'bg-accent']}
				onclick={() => (downloadsOpen = !downloadsOpen)}
			>
				<Download class="h-8 w-8" />
			</button>
			<a
				href="/settings"
				class={['mt-auto p-2 transition-colors', isActive('/settings') && 'bg-accent']}
			>
				<Cog class="h-8 w-8" />
			</a>
		</div>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class={[
				'relative w-full grow overflow-y-auto transition-opacity',
				downloadsOpen && 'opacity-60'
			]}
			onwheel={(e) => {
				if (downloadsOpen) {
					e.preventDefault();
				}
			}}
			ontouchmove={(e) => {
				if (downloadsOpen) {
					e.preventDefault();
				}
			}}
			onclick={() => {
				if (downloadsOpen) {
					downloadsOpen = false;
				}
			}}
		>
			{@render children()}
		</div>
	</div>
{/if}
