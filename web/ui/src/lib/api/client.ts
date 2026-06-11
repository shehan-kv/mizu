import { goto } from '$app/navigation';
import { resolve } from '$app/paths';

export const BASE_URL = '/api/v1/';

export class ApiError extends Error {
	status: number;
	data: unknown;

	constructor(message: string, status: number, data?: unknown) {
		super(message);

		this.name = 'ApiError';
		this.status = status;
		this.data = data;
	}
}

type ApiFetchOptions = RequestInit;

export async function apiFetch<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
	const headers = new Headers(options.headers || {});

	const isFormData = options.body instanceof FormData;

	if (!isFormData && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}

	let response: Response;

	try {
		response = await fetch(`${BASE_URL}${path}`, {
			...options,
			headers
		});
	} catch (err) {
		throw new Error(`Network error: ${err}`);
	}

	// Global auth handling
	if (response.status === 401) {
		await goto(resolve('/sign-in'));

		throw new ApiError('Unauthorized', 401);
	}

	const contentType = response.headers.get('content-type');

	let data: unknown = null;

	if (contentType?.includes('application/json')) {
		data = await response.json();
	} else {
		data = await response.text();
	}

	if (!response.ok) {
		const message =
			typeof data === 'object' && data !== null && 'err' in data && typeof data.err === 'string'
				? data.err
				: response.statusText;

		throw new ApiError(message, response.status, data);
	}

	return data as T;
}
