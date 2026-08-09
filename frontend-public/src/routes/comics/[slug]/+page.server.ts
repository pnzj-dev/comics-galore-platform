import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';
import type { Comment } from '$lib/components/CommentList.svelte';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const token = cookies.get('token');
	const client = getEncoreClient(token);

	const comic = await client.comics.GetComic(params.slug);

	let likeStatus = null;
	if (token && comic.id) {
		try { likeStatus = await client.comics.GetLikeStatus(comic.id); } catch {}
	}

	let related: Awaited<ReturnType<typeof client.comics.GetComic>>[] = [];
	let comments: Comment[] = [];

	if (comic.id) {
		const [relatedRes, commentsRes] = await Promise.all([
			client.comics.ListComics({
				Page: 1, Limit: 4,
				Language: '',
				Search: '',
				Tag: '',
				Sort: '',
				ExcludeMature: ''
			}).then(r => r.comics || []),
			client.comics.ListComments(comic.id),
		]);
		related = relatedRes.filter(c => c.id !== comic.id);
		comments = commentsRes.comments || [];
	}

	return { comic, likeStatus, related, comments };
};
