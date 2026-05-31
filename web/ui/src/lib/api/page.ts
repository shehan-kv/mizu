export interface PaginatedResponse<T> {
	totalCount: number;
	page: number;
	limit: number;
	items: T[];
}
