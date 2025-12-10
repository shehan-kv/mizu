import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIConflictError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export interface LatestVersion {
	id: number;
	version: string;
}
export interface Contract {
	id: number;
	name: string;
	projectName?: string;
	status: string;
	createdAt: Date;
	versions: number;
	numOfRevisions: number;
	acceptedRevisions: number;
	latestVersion: LatestVersion;
	userSignature?: 'signed' | 'rejected';
}

export async function getContracts(
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
		res = await fetch(`/api/v1/contracts?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch contracts: ${err}`);
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

	const payload = (await res.json()) as PaginatedResponse<Contract>;
	return payload;
}

export async function getContractById(id: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/contracts/${id}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch contracts: ${err}`);
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
				throw new APINotFoundError(`Contracts not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as Contract;
	return payload;
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
		res = await fetch(`/api/v1/contracts/project/${projectId}?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch contracts: ${err}`);
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
				throw new APINotFoundError(`Contracts not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<Contract>;
	return payload;
}

export interface ContractVersion {
	id: number;
	createdAt: Date;
	status: string;
	version: string;
	contract: string;
}

export async function getContractVersions(contractId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/contracts/version/${contractId}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch contract versions: ${err}`);
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
				throw new APINotFoundError(`Contract versions not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as ContractVersion[];
	return payload;
}

export interface ContractSignature {
	firstName: string;
	lastName: string;
	signedAt: Date;
	status: string;
	image?: string;
}

export async function getContractSignatures(versionId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/contracts/version/signature/${versionId}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch contract version signatures: ${err}`);
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
				throw new APINotFoundError(`Contract version signatures not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as ContractSignature[];
	return payload;
}

export interface ContractRevisionParams {
	title: string;
	description: string;
}

export async function createContractRevision(
	contractId: number,
	req: ContractRevisionParams,
	signal?: AbortSignal
) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/contracts/revision/${contractId}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(req),
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to create contract revision : ${err}`);
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
}

export async function signVersion(versionId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/contracts/sign/${versionId}`, {
			method: 'POST',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to sign contract version : ${err}`);
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
			case 409:
				throw new APIConflictError(`Already signed or rejected`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function rejectVersion(versionId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/contracts/reject/${versionId}`, {
			method: 'POST',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to reject contract version : ${err}`);
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
			case 409:
				throw new APIConflictError(`Already signed or rejected`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export interface RevisionUser {
	firstName: string;
	lastName: string;
}

export interface ContractRevision {
	id: number;
	contractId: number;
	title: string;
	description: string;
	createdAt: Date;
	updatedAt?: Date;
	status: string;
	reqUser: RevisionUser;
	resUser?: RevisionUser;
}

export interface ContractRevisionQuery {
	q?: string;
	status?: string;
	page: number;
	limit: number;
}

export async function getContractRevisions(
	contractId: number,
	query: ContractRevisionQuery,
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
		res = await fetch(`/api/v1/contracts/revision/${contractId}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch contract version signatures: ${err}`);
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
				throw new APINotFoundError(`Contract version signatures not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<ContractRevision>;
	return payload;
}

export interface ContractCreateParams {
	name: string;
	version: string;
	contract: string;
}

export async function createContract(
	projectId: number,
	req: ContractCreateParams,
	signal?: AbortSignal
) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/contracts/${projectId}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(req),
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to create contract : ${err}`);
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
			case 409:
				throw new APIConflictError(`Already exists`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}
