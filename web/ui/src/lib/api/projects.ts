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

export type ProjectStatus = 'started' | 'paused' | 'cancelled' | 'completed';

export interface Project {
	id: number;
	name: string;
	createdAt: Date;
	status: ProjectStatus;
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
	status?: ProjectStatus,
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

export type ProjectTaskStatus = 'backlog' | 'in-progress' | 'completed';
export type ProjectTaskPriority = 'high' | 'medium' | 'low';

export interface ProjectMember {
	id: number;
	firstName: string;
	lastName: string;
	title?: string;
	image?: string;
	role: string;
}

export interface ProjectTask {
	id: number;
	projectId: number;
	name: string;
	status: ProjectTaskStatus;
	priority: ProjectTaskPriority;
	description: string;
	createdAt: Date;
	estTimeMinutes: number;
	assignees: ProjectMember[];
}

export interface TaskQuery {
	q?: string;
	status?: ProjectTaskStatus;
	priority?: ProjectTaskPriority;
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

export interface ProjectDetails {
	id: number;
	name: string;
	createdAt: Date;
	status: ProjectStatus;
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

export async function getProjectMembers(projectId: number, signal?: AbortSignal) {
	const url = new URLSearchParams();

	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/members`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch project members: ${err}`);
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

	const payload = (await res.json()) as ProjectMember[];
	return payload;
}

export interface TaskMetric {
	key: string;
	value: number;
}

export async function getTaskCompletedMetricsByProject(projectId: number, signal?: AbortSignal) {
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

export interface ProjectMetric {
	key: string;
	value: number;
}

export async function getProjectCreatedMetrics(signal?: AbortSignal) {
	let res: Response;

	try {
		res = await fetch(`/api/v1/projects/metrics/create`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch projects created metrics: ${err}`);
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

	const payload = (await res.json()) as ProjectMetric[];
	return payload;
}

export interface CreateProjectParams {
	name: string;
	status: ProjectStatus;
	members: number[];
}
export async function createProject(req: CreateProjectParams, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/ `, {
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
				throw new APINotFoundError(`Not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function markProjectStarted(projectId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/status/started`, {
			method: 'POST',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to change project status: ${err}`);
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
				throw new APIConflictError(`Already started`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function markProjectPaused(projectId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/status/paused`, {
			method: 'POST',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to change project status: ${err}`);
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
				throw new APIConflictError(`Already paused`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function markProjectCancelled(projectId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/status/cancelled`, {
			method: 'POST',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to change project status: ${err}`);
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
				throw new APIConflictError(`Already cancelled`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function markProjectCompleted(projectId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/status/completed`, {
			method: 'POST',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to change project status: ${err}`);
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
				throw new APIConflictError(`Already completed`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function deleteProject(projectId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}`, {
			method: 'DELETE',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to delete project: ${err}`);
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

export interface ProjectMembersSetParams {
	members: number[];
}
export async function setProjectMembers(
	projectId: number,
	req: ProjectMembersSetParams,
	signal?: AbortSignal
) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/members`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(req),
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to delete project: ${err}`);
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

export interface CreateTaskParams {
	priority: ProjectTaskPriority;
	status: ProjectTaskStatus;
	name: string;
	description: string;
	estimatedTimeMinutes: number;
	assignees: number[];
}

export async function createTask(projectId: number, req: CreateTaskParams, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${projectId}/task`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(req),
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to create task: ${err}`);
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

export async function getProjectTaskAssignees(taskId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/projects/${taskId}/task/assignees`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to create task: ${err}`);
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

	const payload = (await res.json()) as ProjectMember[];
	return payload;
}
