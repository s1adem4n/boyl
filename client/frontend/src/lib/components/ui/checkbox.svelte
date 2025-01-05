<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import Check from '~icons/lucide/check';

	interface CheckboxProps extends HTMLInputAttributes {}

	let { class: className, checked = $bindable(), ...props }: CheckboxProps = $props();
</script>

<div class="relative flex items-center justify-center">
	<input
		type="checkbox"
		class="border-border bg-background checked:border-primary checked:bg-primary h-6 w-6 appearance-none rounded border"
		bind:checked
		{...props}
	/>
	<Check class="pointer-events-none absolute h-4 w-4 text-white" />
</div>
