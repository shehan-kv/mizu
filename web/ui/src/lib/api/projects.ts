import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export interface Project {
	id: number;
	name: string;
	createdAt: Date;
	status: string;
	totalTasks: number;
	tasksCompleted: number;
	totalInvoices: number;
	invoicesPaid: number;
	totalQuotes: number;
}

/**
 * Sends a GET request to retrieve a paginated list of projects
 * assigned to the requesting user.
 */
export async function getProjects(
	q: string,
	page: number,
	limit: number,
	status: string,
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
		res = await fetch(`/api/v1/projects/?${url.toString()}`, {
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

	const payload = (await res.json()) as PaginatedResponse<Project>;
	return payload;
}

export interface ProjectTask {
	id: number;
	projectId: number;
	name: string;
	status: string;
	priority: string;
	description: string;
	createdAt: Date;
	estTimeMinutes: number;
	assignees: {
		id: number;
		firstName: string;
		lastName: string;
		title?: string;
		image?: string;
	}[];
}

export interface TaskQuery {
	q?: string;
	status?: string;
	priority?: string;
	page: number;
	limit: number;
}
export async function getProjectTasks(projectId: number, query: TaskQuery, signal?: AbortSignal) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) url.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) url.set('status', query.status);

	// set "priority" param if priority is truthy
	if (query.priority) url.set('priority', query.priority);

	url.set('page', query.page.toString());
	url.set('limit', query.limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/task?${url.toString()}`, {
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

	const payload = (await res.json()) as PaginatedResponse<ProjectTask>;
	return payload;
}

export interface ProjectMember {
	id: number;
	firstName: string;
	lastName: string;
	title?: string;
	image?: string;
	role: string;
}

export interface ProjectDetails {
	id: number;
	name: string;
	createdAt: Date;
	status: string;
	taskCount: number;
	taskCompletedCount: number;
	invoiceCount: number;
	invoicePaidCount: number;
	quoteCount: number;
	contractCount: number;
	contractSignedCount: number;
	changeReqCount: number;
	changeReqClosedCount: number;
	fileCount: number;
	members: ProjectMember[];
}

export async function getProjectDetails(projectId: number, signal?: AbortSignal) {
	const url = new URLSearchParams();

	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch project: ${err}`);
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
				throw new APINotFoundError(`Project not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as ProjectDetails;
	return payload;
}

export interface TaskMetric {
	key: string;
	value: number;
}

export async function getTaskCompletedCountByProject(projectId: number, signal?: AbortSignal) {
	let res: Response;

	try {
		res = await fetch(`/api/v1/projects/${projectId}/task/metrics/complete`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch completed tasks: ${err}`);
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
				throw new APINotFoundError(`Project not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as TaskMetric[];
	return payload;
}
