import { apiBaseUrl } from '$lib/server/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const base = apiBaseUrl();

	let hello: { message: string; ts: string } | null = null;
	let healthOk = false;

	try {
		const [helloRes, healthRes] = await Promise.all([
			fetch(`${base}/api/hello`),
			fetch(`${base}/health`)
		]);
		if (helloRes.ok) hello = await helloRes.json();
		if (healthRes.ok) healthOk = true;
	} catch {
		// API offline during SSR — page still renders
	}

	return { hello, healthOk };
};
