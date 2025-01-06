<script lang="ts">
	import type { Snippet } from 'svelte';
	import VanillaTilt, { type TiltOptions } from 'vanilla-tilt';

	let tilt: HTMLElement | null = $state(null);
	$effect(() => {
		if (!tilt) return;

		VanillaTilt.init(
			tilt,
			options || {
				glare: true,
				'max-glare': 0.6
			}
		);
	});

	interface TiltProps {
		class?: string;
		options?: TiltOptions;
		children?: Snippet;
	}

	let { class: className, options, children }: TiltProps = $props();
</script>

<div class={['h-full overflow-hidden rounded', className]} bind:this={tilt}>
	{@render children?.()}
</div>
