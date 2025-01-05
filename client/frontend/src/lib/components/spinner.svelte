<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';

	interface SpinnerProps extends HTMLAttributes<HTMLDivElement> {
		text?: string;
	}

	let { text, class: className, ...props }: SpinnerProps = $props();
</script>

<div class="flex w-full flex-col items-center justify-center gap-2">
	<div
		class={['h-32 w-32 animate-spin rounded-full border-b-2 border-t-2 border-gray-900', className]}
		{...props}
	></div>
	{#if text}
		<p>{text}</p>
	{/if}
</div>
