import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export interface InvoiceSummary {
	id: number;
	projectId: number;
	projectName: string;
	isInvoice: boolean;
	issuedAt: Date;
	dueAt?: Date;
	total: Intl.StringNumericLiteral;
	discount: Intl.StringNumericLiteral;
	tax: Intl.StringNumericLiteral;
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

export async function getInvoices(query: InvoiceQuery, signal?: AbortSignal) {
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
		res = await fetch(`/api/v1/invoices/?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch invoices: ${err}`);
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
				throw new APINotFoundError(`Invoices not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<InvoiceSummary>;
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
		throw new NetworkError(`Failed to fetch invoices: ${err}`);
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
				throw new APINotFoundError(`Invoices not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<InvoiceWithStatus>;
	return payload;
}

export interface InvoiceItem {
	id: number;
	description: string;
	qty: Intl.StringNumericLiteral;
	unitPrice: Intl.StringNumericLiteral;
	unitDiscount: Intl.StringNumericLiteral;
	discountType: string;
	unitTax: Intl.StringNumericLiteral;
	taxType: string;
	totalTax: Intl.StringNumericLiteral;
	totalDiscount: Intl.StringNumericLiteral;
	total: Intl.StringNumericLiteral;
}

export interface InvoiceHistoryUser {
	id: number;
	firstName: string;
	lastName: string;
	title?: string;
	image?: string;
	role: string;
}

export interface InvoiceHistory {
	id: number;
	user: InvoiceHistoryUser;
	event: string;
	recoredAt: Date;
	isInvoice: boolean;
	lastStatus?: string;
	newStatus?: string;
}

export interface InvoiceDetails {
	id: number;
	projectId: number;
	projectName: string;
	isInvoice: boolean;
	status: string;
	issuedAt: Date;
	dueAt: Date;
	total: Intl.StringNumericLiteral;
	discount: Intl.StringNumericLiteral;
	tax: Intl.StringNumericLiteral;
	currencyCode: string;
	note?: string;
	items: InvoiceItem[];
	history: InvoiceHistory[];
}

export async function getInvoiceDetails(invoiceId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/invoices/${invoiceId}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch invoice: ${err}`);
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
				throw new APINotFoundError(`Invoice not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as InvoiceDetails;
	return payload;
}
