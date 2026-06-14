import { apiFetch } from './client';
import type { PaginatedResponse } from './page';

export interface Signatory {
	id: string;
	firstName: string;
	lastName: string;
	role: string;
	status: string;
	title?: string;
	image?: string;
	updatedAt: Date;
}

export interface ContractOverview {
	id: string;
	projectId: string;
	name: string;
	status: string;
	memberSignatoryStatus: string;
	signatories: Signatory[];
	createdAt: Date;
	updatedAt: Date;
}

export interface ContractQuery {
	q?: string;
	status?: string;
	page: number;
	limit: number;
}

export async function getContractOverviews(query: ContractQuery, signal?: AbortSignal) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<ContractOverview>>(`contracts?${params.toString()}`, {
		method: 'GET',
		signal
	});
}

export interface Contract {
	id: string;
	projectId: string;
	name: string;
	status: string;
	terms: string;
	memberSignatoryStatus: string;
	signatories: Signatory[];
	createdAt: Date;
	updatedAt: Date;
}

export async function getContract(id: string, signal?: AbortSignal) {
	return apiFetch<Contract>(`contracts/${id}`, {
		method: 'GET',
		signal
	});
}

export async function getContractOverviewsByProject(
	projectId: string,
	query: ContractQuery,
	signal?: AbortSignal
) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<ContractOverview>>(
		`contracts/overviews/project/${projectId}?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export async function getContractOverviewsByMember(
	memberId: string,
	query: ContractQuery,
	signal?: AbortSignal
) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<ContractOverview>>(
		`contracts/overviews/members/${memberId}?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export interface ContractCreateParams {
	name: string;
	terms: string;
	signatoryIds: string[];
}

export async function createContract(
	projectId: string,
	req: ContractCreateParams,
	signal?: AbortSignal
) {
	return apiFetch<void>(`contracts/${projectId}`, {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export async function signContract(contractId: string, signal?: AbortSignal) {
	return apiFetch<void>(`contracts/${contractId}/sign`, {
		method: 'POST',
		signal
	});
}

export async function rejectContract(contractId: string, signal?: AbortSignal) {
	return apiFetch<void>(`contracts/${contractId}/reject`, {
		method: 'POST',
		signal
	});
}

export interface ContractSignatoriesReplaceParams {
	signatories: string[];
}

export async function getContractSignatories(contractId: string, signal?: AbortSignal) {
	return apiFetch<Signatory[]>(`contracts/${contractId}/signatories`, {
		method: 'GET',
		signal
	});
}

export async function replaceContractSignatories(
	contractId: string,
	req: ContractSignatoriesReplaceParams,
	signal?: AbortSignal
) {
	return apiFetch<void>(`contracts/${contractId}/signatories`, {
		method: 'PUT',
		body: JSON.stringify(req),
		signal
	});
}
