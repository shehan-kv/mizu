interface PaginatedResponse<T> {
	count: number;
	page: number;
	limit: number;
	data: T[];
}
