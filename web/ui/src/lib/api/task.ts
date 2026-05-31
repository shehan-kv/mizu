import type { PROJECT_TASK_PRIORITY, PROJECT_TASK_STATUS } from '$lib/constants/project';
import { apiFetch } from './client';
import type { PaginatedResponse } from './page';

export type TaskStatus = (typeof PROJECT_TASK_STATUS)[number];
export type TaskPriority = (typeof PROJECT_TASK_PRIORITY)[number];

export interface TaskAssignee {
	id: string;
	firstName: string;
	lastName: string;
	title?: string;
	image?: string;
	role: string;
}

export interface Task {
	id: string;
	projectId: string;
	name: string;
	status: TaskStatus;
	priority: TaskPriority;
	description: string;
	estimatedMinutes: number;
	createdAt: Date;
	updatedAt: Date;
	assignees: TaskAssignee[];
}

export interface TaskQuery {
	q?: string;
	status?: TaskStatus;
	priority?: TaskPriority;
	page: number;
	limit: number;
}

export async function getTasks(projectId: string, query: TaskQuery, signal?: AbortSignal) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	// set "priority" param if priority is truthy
	if (query.priority) params.set('priority', query.priority);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<Task>>(`tasks/${projectId}?${params.toString()}`, {
		method: 'GET',
		signal
	});
}

export interface TaskMetric {
	key: string;
	value: number;
}

export async function getTaskCompleteCountByProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<TaskMetric[]>(`tasks/${projectId}/completed-count`, {
		method: 'GET',
		signal
	});
}

export interface CreateTaskParams {
	priority: TaskPriority;
	status: TaskStatus;
	name: string;
	description: string;
	estimatedMinutes: number;
	assigneeIds: string[];
}

export async function createTask(projectId: string, req: CreateTaskParams, signal?: AbortSignal) {
	return apiFetch<void>(`tasks/${projectId}`, {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export async function getTaskAssignees(taskId: string, signal?: AbortSignal) {
	return apiFetch<TaskAssignee[]>(`tasks/${taskId}/assignees`, {
		method: 'GET',
		signal
	});
}

export interface TaskAssigneesReplaceParams {
	assigneeIds: string[];
}
export async function replaceTaskAssignees(
	taskId: string,
	req: TaskAssigneesReplaceParams,
	signal?: AbortSignal
) {
	return apiFetch<void>(`tasks/${taskId}/assignees`, {
		method: 'PUT',
		body: JSON.stringify(req),
		signal
	});
}
