import type { Base } from './types';
import type { RecordService } from 'pocketbase';

export function subscribeSingle<T extends Base>(
	service: RecordService<T>,
	setState: (state: T | null) => void,
	query: string
) {
	service.getOne(query).then((res) => {
		setState(res);
	});

	const unsubscribe = service.subscribe(query, (e) => {
		switch (e.action) {
			case 'create':
				setState(e.record);
				break;
			case 'update':
				setState(e.record);
				break;
			case 'delete':
				setState(null);
				break;
		}
	});

	return () => unsubscribe.then((fn) => fn());
}

export async function subscribeMultiple<T extends Base>(
	service: RecordService<T>,
	getState: () => T[],
	setState: (state: T[]) => void,
	query: string
) {
	await service.getList(1, 1000).then((res) => {
		setState(res.items);
	});

	const unsubscribe = await service.subscribe(query, (e) => {
		switch (e.action) {
			case 'create':
				setState(getState().concat(e.record));
				break;
			case 'update':
				setState(getState().map((r) => (r.id === e.record.id ? e.record : r)));
				break;
			case 'delete':
				setState(getState().filter((r) => r.id !== e.record.id));
				break;
		}
	});

	return unsubscribe;
}

export function validatePath(path: string, style: 'windows' | 'unix'): boolean {
	const windowsPattern = /^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]*$/;
	const unixPattern = /^(\/[^/]+)+\/?$/;

	if (style === 'windows') {
		return windowsPattern.test(path);
	} else if (style === 'unix') {
		return unixPattern.test(path);
	} else {
		throw new Error("Invalid style specified. Use 'windows' or 'unix'.");
	}
}

export function validateUrl(url: string): boolean {
	const pattern = /^(http|https):\/\/[^ "]+$/;
	return pattern.test(url);
}

export function validateEmail(email: string): boolean {
	const pattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	return pattern.test(email);
}

export function formatBytes(bytes: number, decimals = 2): string {
	if (bytes === 0 || isNaN(bytes)) return '0 Bytes';

	const k = 1024;
	const dm = decimals < 0 ? 0 : decimals;
	const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

	const i = Math.floor(Math.log(bytes) / Math.log(k));

	return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export function sortCreated<T extends Base>(list: T[]): T[] {
	return list.sort((a, b) => new Date(b.created).getTime() - new Date(a.created).getTime());
}
