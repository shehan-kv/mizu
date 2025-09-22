import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	APIUnauthorizedError,
	NetworkError
} from './errors';

interface FileUser {
	id: number;
	firstName: string;
	lastName: string;
}

export interface File {
	id: number;
	channelId: number;
	user: FileUser;
	originalName: string;
	savedName: string;
	uploadedAt: Date;
	url: string;
	size: number;
}

/**
 * Sends a GET request to retrieve a paginated list of files
 * for the specified channel.
 */
export async function getFilesByChannel(
	channelId: number,
	q: string,
	page: number,
	limit: number,
	signal?: AbortSignal
) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (q) url.set('q', q);

	url.set('page', page.toString());
	url.set('limit', limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/files/${channelId}?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch files: ${err}`);
	}

	if (!res.ok) {
		switch (res.status) {
			case 400:
				throw new APIBadRequestError('Bad request');
			case 401:
				throw new APIUnauthorizedError('Unauthorized');
			case 403:
				throw new APIForbiddenError('Forbidden');
			case 404:
				throw new APINotFoundError(`Channel ${channelId} not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<File>;
	return payload;
}
