import type { Actions } from './$types';

export const actions: Actions = {
	signIn: async ({ locals, url }) => {
		await locals.logtoClient.signIn({ redirectUri: `${url.origin}/callback`, firstScreen: 'signIn' });
	},
	forgotPassword: async ({ locals, url }) => {
		await locals.logtoClient.signIn({ redirectUri: `${url.origin}/callback`, firstScreen: 'reset_password' });
	},
};
