interface SignInParams {
	email: string;
	password: string;
	rememberMe: boolean;
}

/**
 * Sends a POST request to the sign-in API endpoint with the user's credentials.
 */
export function signIn(creds: SignInParams) {
	return fetch('/api/v1/auth/sign-in', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(creds)
	});
}
