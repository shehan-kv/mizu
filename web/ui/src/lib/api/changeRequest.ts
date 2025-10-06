import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export interface ChangeRequestUser {
	id: string;
	firstName: string;
	lastName: string;
	title?: string;
	role?: string;
	image?: string;
}

export interface ChangeRequest {
	id: number;
	title: string;
	createdAt: Date;
	status: string;
	requestedBy: ChangeRequestUser;
	project: {
		id: number;
		name: string;
	};
}

export interface ChangeRequestQuery {
	q?: string;
	status?: string;
	page: number;
	limit: number;
}

export async function getChangeRequests(
	q: string,
	status: string,
	page: number,
	limit: number,
	signal?: AbortSignal
) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (q) url.set('q', q);

	// set "status" param if status is truthy
	if (status) url.set('status', status);

	url.set('page', page.toString());
	url.set('limit', limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/change-requests/?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch change requests: ${err}`);
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
				throw new APINotFoundError(`Change requests not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<ChangeRequest>;
	return payload;
}

export async function getChangeRequestsByProject(
	projectId: number,
	query: ChangeRequestQuery,
	signal?: AbortSignal
) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) url.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) url.set('status', query.status);

	url.set('page', query.page.toString());
	url.set('limit', query.limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/change-requests/project/${projectId}?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch change requests: ${err}`);
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
				throw new APINotFoundError(`Change requests not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<ChangeRequest>;
	return payload;
}

export interface CreateChangeRequestParams {
	title: string;
	content: string;
}

export async function createChangeRequest(
	projectId: number,
	req: CreateChangeRequestParams,
	signal?: AbortSignal
) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/change-requests/${projectId}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(req),
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch projects: ${err}`);
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
				throw new APINotFoundError(`Change requests not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export interface ChangeRequestEntry {
	id: number;
	user: ChangeRequestUser;
	createdAt: Date;
	content: string;
}

export interface ChangeRequestDetails extends ChangeRequest {
	entries: ChangeRequestEntry[];
}

export async function getChangeRequestDetails(requestId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/change-requests/${requestId}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch change requests: ${err}`);
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
				throw new APINotFoundError(`Change requests not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as ChangeRequestDetails;
	return payload;
}
