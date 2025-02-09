import { subscribeMultiple } from './utils';
import client from './client';
import server from './remote';
import type { ClientGame, ClientSettings, Download, Game, Setting, User } from './types';

class ClientState {
	#rawSettings: Setting[] = $state([]);
	settings: ClientSettings = $derived.by(() => {
		const settings: ClientSettings = {
			setup: false,
			os: 'windows',
			gamesDirectory: '',
			defaultLauncher: '',
			serverUrl: '',
			email: '',
			password: '',
			fitCovers: true
		};
		this.#rawSettings.forEach((setting) => {
			settings[setting.key] = setting.value;
		});

		return settings;
	});
	downloads: Download[] = $state([]);
	games: ClientGame[] = $state([]);

	async load() {
		const settingsUnsubscribe = await subscribeMultiple(
			client.collection('settings'),
			() => this.#rawSettings,
			(settings) => (this.#rawSettings = settings),
			'*'
		);
		const downloadsUnsubscribe = await subscribeMultiple(
			client.collection('downloads'),
			() => this.downloads,
			(downloads) => (this.downloads = downloads),
			'*'
		);
		const gamesUnsubscribe = await subscribeMultiple(
			client.collection('games'),
			() => this.games,
			(games) => (this.games = games),
			'*'
		);

		return () => {
			settingsUnsubscribe();
			downloadsUnsubscribe();
			gamesUnsubscribe();
		};
	}

	async setSetting(key: string, value: unknown) {
		console.log('setSetting', key, value);
		const existing = this.#rawSettings.find((setting) => setting.key === key);
		if (existing) {
			await client.collection('settings').update(existing.id, {
				value: JSON.stringify(value)
			});
		} else {
			await client.collection('settings').create({ key, value: JSON.stringify(value) });
		}
	}

	async addDownload(id: string) {
		const res = await client.send(`/api/download?id=${id}`, {
			method: 'POST'
		});
		if (!res.ok) {
			throw new Error('Failed to download game');
		}
	}

	async cancelDownload(id: string) {
		const res = await client.send(`/api/download?id=${id}`, {
			method: 'DELETE'
		});

		if (!res.ok) {
			throw new Error('Failed to cancel download');
		}
	}

	async launchGame(id: string) {
		const res = await client.send(`/api/launch?id=${id}`, {
			method: 'POST'
		});

		if (!res.ok) {
			throw new Error('Failed to launch game');
		}
	}
}

export const clientState = new ClientState();

class RemoteState {
	games: Game[] = $state([]);
	auth: { token: string; record: User | null } = $state({ token: '', record: null });

	async load() {
		server.authStore.onChange((token, record) => {
			this.auth = { token, record: record as unknown as User };
		});
		this.auth = {
			token: server.authStore.token,
			record: server.authStore.record as unknown as User
		};

		const gamesUnsubscribe = await subscribeMultiple(
			server.collection('games'),
			() => this.games,
			(state) => (this.games = state),
			'*'
		);

		return () => {
			gamesUnsubscribe();
		};
	}
}

export const remoteState = new RemoteState();
