import { BASE_URL } from '$lib/api/client';
import type { PageLoad } from './$types';

export const csr = true;

export const load: PageLoad = async ({ params, fetch }) => {
	const id = params.id;

	const validatePromise = fetch(`${BASE_URL}users/verifications/${id}`).then((res) => {
		if (!res.ok) {
			return {
				id,
				ok: false,
				loadError: {
					status: res.status,
					message: 'Invalid Request'
				}
			};
		}

		return {
			id,
			ok: true,
			loadError: null
		};
	});

	return {
		id,
		validatePromise
	};
};
