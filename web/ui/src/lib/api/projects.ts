import type { PROJECT_STATUS } from '$lib/constants/project';
import { apiFetch } from './client';
import type { PaginatedResponse } from './page';

export type ProjectStatus = (typeof PROJECT_STATUS)[number];

export interface ProjectStat {
	id: string;
	name: string;
	status: ProjectStatus;
	createdAt: Date;
	totalTasks: number;
	tasksCompleted: number;
	totalInvoices: number;
	invoicesPaid: number;
	totalQuotes: number;
}

export interface ProjectQuery {
	q?: string;
	status?: string;
	page: number;
	limit: number;
}

export async function getProjectStats(query: ProjectQuery, signal?: AbortSignal) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) params.set('q', query.q);

	// set "status" param if status is truthy
	if (query.status) params.set('status', query.status);

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<ProjectStat>>(`projects/stats?${params.toString()}`, {
		method: 'GET',
		signal
	});
}

export async function getProjectStatsByMember(
	memberId: string,
	req: ProjectQuery,
	signal?: AbortSignal
) {
	const params = new URLSearchParams();

	// set "q" param if q is truthy
	if (req.q) params.set('q', req.q);

	// set "status" param if status is truthy
	if (req.status) params.set('status', req.status);

	params.set('page', req.page.toString());
	params.set('limit', req.limit.toString());

	return apiFetch<PaginatedResponse<ProjectStat>>(
		`projects/stats/members/${memberId}?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export async function getAllProjectStatsByMember(memberId: string, signal?: AbortSignal) {
	return apiFetch<ProjectStat[]>(`projects/stats/members/${memberId}/all`, {
		method: 'GET',
		signal
	});
}

export interface ProjectMember {
	id: string;
	firstName: string;
	lastName: string;
	title?: string;
	image?: string;
	role: string;
}

export interface Project {
	id: string;
	name: string;
	status: ProjectStatus;
	createdAt: Date;
	members: ProjectMember[];
	taskCount: number;
	taskCompletedCount: number;
	invoiceCount: number;
	invoicePaidCount: number;
	quoteCount: number;
	contractCount: number;
	contractSignedCount: number;
	fileCount: number;
}

export async function getProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<Project>(`projects/${projectId}`, {
		method: 'GET',
		signal
	});
}

export async function getProjectMembers(projectId: string, q: string, signal?: AbortSignal) {
	const params = new URLSearchParams();
	if (q) params.set('q', q);

	return apiFetch<ProjectMember[]>(`projects/${projectId}/members?${params.toString()}`, {
		method: 'GET',
		signal
	});
}

export interface ProjectMetric {
	key: string;
	value: number;
}

export async function getProjectCreatedMetrics(signal?: AbortSignal) {
	return apiFetch<ProjectMetric[]>(`projects/stats/created-count`, {
		method: 'GET',
		signal
	});
}

export interface CreateProjectParams {
	name: string;
	status: ProjectStatus;
	members: string[];
}
export async function createProject(req: CreateProjectParams, signal?: AbortSignal) {
	return apiFetch<void>(`projects`, {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export async function startProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<void>(`projects/${projectId}/start`, {
		method: 'PUT',
		signal
	});
}

export async function pauseProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<void>(`projects/${projectId}/pause`, {
		method: 'PUT',
		signal
	});
}

export async function cancelProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<void>(`projects/${projectId}/cancel`, {
		method: 'PUT',
		signal
	});
}

export async function completeProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<void>(`projects/${projectId}/complete`, {
		method: 'PUT',
		signal
	});
}

export async function deleteProject(projectId: string, signal?: AbortSignal) {
	return apiFetch<void>(`projects/${projectId}`, {
		method: 'DELETE',
		signal
	});
}

export interface ProjectMembersReplaceParams {
	memberIds: string[];
}
export async function replaceProjectMembers(
	projectId: string,
	req: ProjectMembersReplaceParams,
	signal?: AbortSignal
) {
	return apiFetch<void>(`projects/${projectId}/members`, {
		method: 'PUT',
		body: JSON.stringify(req),
		signal
	});
}

export interface MemberProjectsReplaceParams {
	projectIds: string[];
}

export async function replaceMemberProjects(
	memberId: string,
	req: MemberProjectsReplaceParams,
	signal?: AbortSignal
) {
	return apiFetch<void>(`projects/members/${memberId}/replace`, {
		method: 'PUT',
		body: JSON.stringify(req),
		signal
	});
}
