<script lang="ts">
	import { Dropdown, DropdownItem } from '$lib/components/ui/dropdown';
	import { Button } from '$lib/components/ui';

	let open = $state(false);
</script>

<div class="flex flex-col gap-4 p-4">
	<Dropdown bind:open>
		{#snippet trigger()}
			<Button onclick={() => (open = !open)}>Toggle</Button>
		{/snippet}

		{#snippet content()}
			<DropdownItem>Item 1</DropdownItem>
			<DropdownItem>Item 1</DropdownItem>
			<DropdownItem>Item 1</DropdownItem>
			<DropdownItem>Item 1</DropdownItem>
		{/snippet}
	</Dropdown>
</div>
