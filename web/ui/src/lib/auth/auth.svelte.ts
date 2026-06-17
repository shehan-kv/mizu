import { getMe, type UserRole } from '$lib/api/users';

class Auth {
	role = $state<UserRole | null>(null);
	loading = $state(true);
	error = $state<unknown | null>(null);

	async init() {
		this.loading = true;
		this.error = null;

		try {
			const res = await getMe();
			this.role = res.role;
		} catch (e) {
			// API function already redirects to /sign-in on 401

			this.error = e;
			this.role = null;
		} finally {
			this.loading = false;
		}
	}

	clear() {
		this.role = null;
		this.loading = false;
		this.error = null;
	}
}

export const auth = new Auth();
