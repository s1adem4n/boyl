import PocketBase, { type RecordService } from 'pocketbase';
import type { ClientGame, Download, Setting } from './types';

interface TypedPocketBase extends PocketBase {
	collection(idOrName: string): RecordService;
	collection(idOrName: 'settings'): RecordService<Setting>;
	collection(idOrName: 'downloads'): RecordService<Download>;
	collection(idOrName: 'games'): RecordService<ClientGame>;
}

const client = new PocketBase('http://localhost:48658') as TypedPocketBase;
client.autoCancellation(false);

export default client;
