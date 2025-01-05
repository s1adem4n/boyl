<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';

	interface ButtonProps extends HTMLButtonAttributes {}

	let { class: className, children, ...props }: ButtonProps = $props();
</script>

<button
	class={[
		'bg-primary text-primary-foreground hover:opacity-blur flex h-9 w-full items-center justify-center rounded py-2 transition-opacity disabled:opacity-50',
		className
	]}
	{...props}
>
	{@render children?.()}
</button>
