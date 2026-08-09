export interface JWTUser {
	id: string;
	email: string;
	role: string;
	tier: string;
	created_at: string;
}

export function decodeJWT(token: string): JWTUser | null {
	try {
		const payload = JSON.parse(atob(token.split('.')[1]));
		return {
			id: payload.UserID || payload.user_id || '',
			email: payload.Email || payload.email || '',
			role: payload.Role || payload.role || 'user',
			tier: payload.Tier || payload.tier || 'free',
			created_at: '',
		};
	} catch {
		return null;
	}
}
