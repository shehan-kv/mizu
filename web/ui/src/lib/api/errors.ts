export class NetworkError extends Error {}

export class APIError extends Error {
	httpStatus: number;
	constructor(message: string, httpStatus: number) {
		super(message);
		this.name = 'APIError';
		this.httpStatus = httpStatus;
	}
}

export class APIBadRequestError extends APIError {
	constructor(message: string) {
		super(message, 400);
		this.name = 'APIBadRequestError';
	}
}

export class APIUnauthorizedError extends APIError {
	constructor(message: string) {
		super(message, 401);
		this.name = 'APIUnauthorizedError';
	}
}

export class APIForbiddenError extends APIError {
	constructor(message: string) {
		super(message, 403);
		this.name = 'APIForbiddenError';
	}
}

export class APIServerError extends APIError {
	constructor(message: string) {
		super(message, 500);
		this.name = 'APIServerError';
	}
}

export class APINotFoundError extends APIError {
	constructor(message: string) {
		super(message, 404);
		this.name = 'APINotFoundError';
	}
}
