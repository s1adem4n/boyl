export interface Base {
	id: string;
	created: string;
	updated: string;
}

export interface Game extends Base {
	path: string;
	name: string;
	status: 'deleted' | 'invalid' | 'missing' | 'found';
	version: string;
	summary: string;
	released: string;
	rating: number;
	genres: string[];
	cover: string;
	artworks: string[];
	screenshots: string[];
	provider: string;
	providerId: string;
}

export interface ClientGame extends Base {
	game: string;
	path: string;
	executable: string;
}

export interface Status {
	name: string;
	text: string;
	current: number;
	total: number;
}

export interface Download extends Base {
	game: string;
	status: 'starting' | 'downloading' | 'extracting' | 'completed' | 'failed';
	active: boolean;
	text: string;
	speed: number;
	progress: number;
	total: number;
}

export interface Setting extends Base {
	key: string;
	value: string;
}

export interface ClientSettings {
	[key: string]: unknown;
	os: 'windows' | 'linux' | 'darwin';
	gamesDirectory: string;
	setup: boolean;
	serverUrl: string;
	email: string;
	password: string;
	fitCovers: boolean;
}
