import { env } from '$env/dynamic/private';
import { env as publicEnv } from '$env/dynamic/public';

export function apiBaseUrl(): string {
	return (
		env.API_INTERNAL_URL ||
		publicEnv.PUBLIC_API_URL ||
		'http://127.0.0.1:3000'
	);
}
