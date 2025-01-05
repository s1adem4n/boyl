<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button, Input } from '$lib/components/ui';
	import remote from '$lib/remote';
	import client from '$lib/client';
	import { clientState } from '$lib/state.svelte';
	import { validateEmail, validatePath, validateUrl } from '$lib/utils';
	import { ClientResponseError } from 'pocketbase';

	let gamesDirectory = $state(clientState.settings.gamesDirectory);
	let serverUrl = $state(clientState.settings.serverUrl);
	let email = $state(clientState.settings.email);
	let password = $state(clientState.settings.password);
	let error = $state('');

	let valid = $derived(
		validatePath(gamesDirectory, clientState.settings.os === 'windows' ? 'windows' : 'unix') &&
			validateUrl(serverUrl) &&
			validateEmail(email) &&
			password.length > 0
	);
</script>

<div class="flex flex-col gap-4 p-4">
	<h1 class="text-2xl font-bold">Setup</h1>

	<h2 class="text-xl font-bold">Client settings</h2>

	<div class="flex flex-col gap-2">
		<p class="text-sm text-gray-400">Where should your games be saved?</p>
		<label class="flex flex-col gap-1">
			Directory
			<Input
				type="text"
				placeholder={clientState.settings.os === 'windows' ? 'C:\\Games' : '/home/yourname/Games'}
				bind:value={gamesDirectory}
			/>
			{#if !validatePath(gamesDirectory, clientState.settings.os === 'windows' ? 'windows' : 'unix')}
				<p class="text-sm text-red-500">Please enter a valid path</p>
			{/if}
		</label>
	</div>

	<h2 class="text-xl font-bold">Server settings</h2>
	<div class="flex flex-col gap-2">
		<p class="text-sm text-gray-400">Please setup the connection to your server.</p>
		<label class="flex flex-col gap-1">
			URL
			<Input type="text" placeholder="https://yourserver.domain.com" bind:value={serverUrl} />
			{#if !validateUrl(serverUrl)}
				<p class="text-sm text-red-500">Please enter a valid URL</p>
			{/if}
		</label>
	</div>

	<div class="flex flex-col gap-2">
		<label class="flex flex-col gap-1">
			Email
			<Input type="text" placeholder="example@email.com" bind:value={email} />
			{#if !validateEmail(email)}
				<p class="text-sm text-red-500">Please enter a valid email</p>
			{/if}
		</label>
		<label class="flex flex-col gap-1">
			Password
			<Input type="password" placeholder="*****" bind:value={password} />
			{#if password.length === 0}
				<p class="text-sm text-red-500">Please enter a password</p>
			{/if}
		</label>
	</div>

	<Button
		disabled={!valid}
		onclick={async () => {
			clientState.setSetting('setup', 'true');
			clientState.setSetting('gamesDirectory', gamesDirectory);
			clientState.setSetting('serverUrl', serverUrl);
			clientState.setSetting('email', email);
			clientState.setSetting('password', password);

			remote.baseURL = serverUrl;
			try {
				await remote.collection('users').authWithPassword(email, password);
				goto('/');
			} catch (e) {
				if (e instanceof ClientResponseError) {
					error = e.message;
				}
			}
			await client.send('/api/update-remote', {});
		}}
	>
		Save
	</Button>
	{#if error}
		<p class="text-sm text-red-500">{error}</p>
	{/if}
</div>
