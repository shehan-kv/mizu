import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export interface Contract {
	id: number;
	name: string;
	status: string;
	createdAt: Date;
	versions: number;
	numOfRevisions: number;
	acceptedRevisions: number;
}

export async function getContractsByProject(
	projectId: number,
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
		res = await fetch(`/api/v1/contracts/${projectId}?${url.toString()}`, {
			method: 'GET',
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
				throw new APINotFoundError(`Projects not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<Contract>;
	return payload;
}
