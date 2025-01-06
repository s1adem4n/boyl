import PocketBase, { type RecordService } from 'pocketbase';
import type { Game, Status, User } from './types';

interface TypedPocketBase extends PocketBase {
	collection(idOrName: string): RecordService;
	collection(idOrName: 'users'): RecordService<User>;
	collection(idOrName: 'games'): RecordService<Game>;
	collection(idOrName: 'status'): RecordService<Status>;
}

const remote = new PocketBase('http://localhost:8091') as TypedPocketBase;

export default remote;
