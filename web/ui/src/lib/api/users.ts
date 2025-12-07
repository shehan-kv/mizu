import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export type UserRole = 'admin' | 'staff' | 'client';

export interface User {
	id: number;
	firstName: string;
	lastName: string;
	title?: string;
	email: string;
	role: UserRole;
	image?: string;
	createdAt: Date;
	lastLogin?: Date;
	isActive: boolean;
}

export interface UserQuery {
	q?: string;
	role?: string;
	page: number;
	limit: number;
}
export async function getUsers(query: UserQuery, signal?: AbortSignal) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) url.set('q', query.q);

	// set "role" param if status is truthy
	if (query.role) url.set('role', query.role);

	url.set('page', query.page.toString());
	url.set('limit', query.limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/users/?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch users: ${err}`);
	}

	if (!res.ok) {
		switch (res.status) {
			case 400:
				throw new APIBadRequestError('Bad request');
			case 401:
				goto('/sign-in');
			case 403:
				throw new APIForbiddenError('Forbidden');
			case 404:
				throw new APINotFoundError(`Not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<User>;
	return payload;
}
