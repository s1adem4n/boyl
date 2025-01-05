<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface InputProps extends HTMLInputAttributes {}

	let { class: className, value = $bindable(), ...props }: InputProps = $props();
</script>

<input
	class={[
		'border-border focus-visible:border-primary rounded border px-2 py-1 focus-visible:outline-none',
		className
	]}
	bind:value
	{...props}
/>
