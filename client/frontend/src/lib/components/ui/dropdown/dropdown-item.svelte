<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';

	interface ButtonProps extends HTMLButtonAttributes {}

	let { class: className, children, ...props }: ButtonProps = $props();
</script>

<button
	class={['hover:bg-accent flex items-center gap-2 p-2 transition-colors', className]}
	{...props}
>
	{@render children?.()}
</button>
