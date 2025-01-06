<script lang="ts">
	import { flip, type Placement } from '@floating-ui/dom';
	import { offset } from '@floating-ui/dom';
	import { autoUpdate } from '@floating-ui/dom';
	import { shift } from '@floating-ui/dom';
	import { computePosition } from '@floating-ui/dom';
	import type { Snippet } from 'svelte';
	import { fade } from 'svelte/transition';

	interface DropdownProps {
		trigger: Snippet;
		content: Snippet;
		placement?: Placement;
		open?: boolean;
	}

	let { trigger, content, placement, open = $bindable() }: DropdownProps = $props();

	let triggerElement: HTMLElement | null = $state(null);
	let contentElement: HTMLElement | null = $state(null);

	function updatePosition() {
		if (!triggerElement || !contentElement) return;

		computePosition(triggerElement, contentElement, {
			placement,
			middleware: [flip(), shift({ padding: 8 }), offset(8)]
		}).then(({ x, y }) => {
			if (!contentElement) return;
			Object.assign(contentElement.style, {
				top: `${y}px`,
				left: `${x}px`
			});
		});
	}

	$effect(() => {
		if (!triggerElement || !contentElement) return;
		const cleanUp = autoUpdate(triggerElement, contentElement, updatePosition);

		return cleanUp;
	});
</script>

<div>
	<div bind:this={triggerElement}>
		{@render trigger?.()}
	</div>
	{#if open}
		<div
			transition:fade={{ duration: 100 }}
			class="border-border divide-border bg-background absolute flex flex-col divide-y overflow-hidden rounded border"
			bind:this={contentElement}
		>
			{@render content?.()}
		</div>
	{/if}
</div>
