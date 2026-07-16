import { env as publicEnv } from '$env/dynamic/public';

export function publicApiUrl(): string {
	return publicEnv.PUBLIC_API_URL || 'http://127.0.0.1:3000';
}

export function wsUrl(): string {
	if (typeof window === 'undefined') {
		return publicEnv.PUBLIC_WS_URL || 'ws://127.0.0.1:3000/ws/metrics';
	}
	const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	return publicEnv.PUBLIC_WS_URL || `${proto}//${window.location.host}/ws/metrics`;
}
