import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export interface Invoice {
	id: number;
	projectId: number;
	projectName: string;
	isInvoice: boolean;
	issuedAt: Date;
	dueAt?: Date;
	total: Intl.StringNumericLiteral;
	currencyCode: string;
	status: string;
}

export interface InvoiceQuery {
	q?: string;
	status?: string;
	type?: string;
	page: number;
	limit: number;
}

export async function getInvoices(
	q: string,
	status: string,
	type: string,
	page: number,
	limit: number,
	signal?: AbortSignal
) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (q) url.set('q', q);

	// set "status" param if status is truthy
	if (status) url.set('status', status);

	// set "type" param if type is truthy
	if (type) url.set('type', type);

	url.set('page', page.toString());
	url.set('limit', limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/invoices/?${url.toString()}`, {
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

	const payload = (await res.json()) as PaginatedResponse<Invoice>;
	return payload;
}

export interface InvoiceWithStatus {
	id: number;
	isInvoice: boolean;
	issuedAt: Date;
	dueAt?: Date;
	total: Intl.StringNumericLiteral;
	currencyCode: string;
	status: string;
}

export async function getInvoicesByProjectId(
	projectId: number,
	query: InvoiceQuery,
	signal?: AbortSignal
) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) url.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) url.set('status', query.status);

	// set "type" param if type is truthy
	if (query.type) url.set('type', query.type);

	url.set('page', query.page.toString());
	url.set('limit', query.limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/invoices/project/${projectId}?${url.toString()}`, {
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

	const payload = (await res.json()) as PaginatedResponse<InvoiceWithStatus>;
	return payload;
}
