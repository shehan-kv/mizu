import type { INVOICE_STATUS } from '$lib/constants/invoice';
import { apiFetch } from './client';
import type { PaginatedResponse } from './page';

export type InvoiceStatus = (typeof INVOICE_STATUS)[number];

export interface InvoiceOverview {
	id: string;
	projectId: string;
	projectName: string;
	isInvoice: boolean;
	status: InvoiceStatus;
	dueAt?: Date;
	currencyCode: string;
	note?: string;
	totalTax: Intl.StringNumericLiteral;
	totalDiscount: Intl.StringNumericLiteral;
	subTotal: Intl.StringNumericLiteral;
	createdAt: Date;
	updatedAt: Date;
}

export interface InvoiceQuery {
	q?: string;
	status?: string;
	type?: string;
	page: number;
	limit: number;
}

export async function getInvoices(query: InvoiceQuery, signal?: AbortSignal) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	// set "type" param if type is truthy
	if (query.type) params.set('type', query.type);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<InvoiceOverview>>(`invoices?${params.toString()}`, {
		method: 'GET',
		signal
	});
}

export async function getInvoicesByProject(
	projectId: string,
	query: InvoiceQuery,
	signal?: AbortSignal
) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	// set "type" param if type is truthy
	if (query.type) params.set('type', query.type);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<InvoiceOverview>>(
		`invoices/project/${projectId}?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export async function getInvoicesByMember(
	memberId: string,
	query: InvoiceQuery,
	signal?: AbortSignal
) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	// set "type" param if type is truthy
	if (query.type) params.set('type', query.type);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<InvoiceOverview>>(
		`invoices/member/${memberId}?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export interface InvoiceItem {
	id: string;
	description: string;
	qty: Intl.StringNumericLiteral;
	unitPrice: Intl.StringNumericLiteral;
	discountRate: Intl.StringNumericLiteral;
	discountType: string;
	taxRate: Intl.StringNumericLiteral;
	taxType: string;
	discountAmountPerUnit: Intl.StringNumericLiteral;
	taxAmountPerUnit: Intl.StringNumericLiteral;
	taxableBasePerUnit: Intl.StringNumericLiteral;
	lineGross: Intl.StringNumericLiteral;
	lineDiscount: Intl.StringNumericLiteral;
	lineNet: Intl.StringNumericLiteral;
	lineTax: Intl.StringNumericLiteral;
	lineTotal: Intl.StringNumericLiteral;
}

export interface Invoice {
	id: string;
	projectId: string;
	projectName: string;
	isInvoice: boolean;
	status: InvoiceStatus;
	dueAt?: Date;
	currencyCode: string;
	note?: string;
	totalTax: Intl.StringNumericLiteral;
	totalDiscount: Intl.StringNumericLiteral;
	subTotal: Intl.StringNumericLiteral;
	items: InvoiceItem[];
	createdAt: Date;
	updatedAt: Date;
}

export async function getInvoice(invoiceId: string, signal?: AbortSignal) {
	return apiFetch<Invoice>(`invoices/${invoiceId}`, {
		method: 'GET',
		signal
	});
}

export async function acceptInvoice(invoiceId: string, signal?: AbortSignal) {
	return apiFetch<void>(`invoices/${invoiceId}/accept`, {
		method: 'PUT',
		signal
	});
}

export async function rejectInvoice(invoiceId: string, signal?: AbortSignal) {
	return apiFetch<void>(`invoices/${invoiceId}/reject`, {
		method: 'PUT',
		signal
	});
}

export async function payInvoice(invoiceId: string, signal?: AbortSignal) {
	return apiFetch<void>(`invoices/${invoiceId}/pay`, {
		method: 'PUT',
		signal
	});
}

export async function cancelInvoice(invoiceId: string, signal?: AbortSignal) {
	return apiFetch<void>(`invoices/${invoiceId}/cancel`, {
		method: 'PUT',
		signal
	});
}

export async function convertInvoice(invoiceId: string, signal?: AbortSignal) {
	return apiFetch<void>(`invoices/${invoiceId}/convert`, {
		method: 'PUT',
		signal
	});
}

export interface InvoiceMetric {
	key: string;
	value: number;
}

export async function getPaidInvoiceCountByProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<InvoiceMetric[]>(`invoices/paid-count/project/${projectId}`, {
		method: 'GET',
		signal
	});
}

export async function getPaidInvoiceCountByMember(memberId: string, signal?: AbortSignal) {
	return apiFetch<InvoiceMetric[]>(`invoices/paid-count/member/${memberId}`, {
		method: 'GET',
		signal
	});
}

export async function getPaidInvoiceCount(signal?: AbortSignal) {
	return apiFetch<InvoiceMetric[]>(`invoices/paid-count`, {
		method: 'GET',
		signal
	});
}

export interface InvoicesSummaryMetric {
	currencyCode: string;
	count: number;
	amount: Intl.StringNumericLiteral;
}

export interface InvoicesSummary {
	invoicesPaid: InvoicesSummaryMetric[];
	invoicesPending: InvoicesSummaryMetric[];
	invoicesAccepted: InvoicesSummaryMetric[];
	invoicesRejected: InvoicesSummaryMetric[];
	invoicesCancelled: InvoicesSummaryMetric[];
	quotesPending: InvoicesSummaryMetric[];
	quotesRejected: InvoicesSummaryMetric[];
}

export async function getInvoiceSummmaryByMember(memberId: string, signal?: AbortSignal) {
	return apiFetch<InvoiceOverview>(`invoices/member/${memberId}/summary`, {
		method: 'GET',
		signal
	});
}

export async function getInvoicesSummmary(signal?: AbortSignal) {
	return apiFetch<InvoicesSummary>(`invoices/summary`, {
		method: 'GET',
		signal
	});
}

export interface InvoiceItemParams {
	description: string;
	qty: string;
	unitPrice: string;
	discountRate: string;
	discountType: string;
	taxRate: string;
	taxType: string;
}

export interface CreateInvoiceParams {
	isInvoice: boolean;
	currencyCode: string;
	note: string;
	dueDate?: Date;
	items: InvoiceItemParams[];
}
export async function createInvoice(
	projectId: string,
	req: CreateInvoiceParams,
	signal: AbortSignal
) {
	return apiFetch<void>(`invoices/${projectId}`, {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export async function emailInvoice(invoiceId: string, signal?: AbortSignal) {
	return apiFetch<void>(`invoices/${invoiceId}/email`, {
		method: 'POST',
		signal
	});
}
