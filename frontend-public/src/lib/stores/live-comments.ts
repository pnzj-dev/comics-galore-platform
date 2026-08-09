const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:4000';

type CommentHandler = (data: unknown) => void;

export function createCommentStream(comicId: string) {
	const url = `${API_BASE}/comments-stream/${comicId}`;
	const es = new EventSource(url);
	const handlers = new Set<CommentHandler>();

	es.onmessage = (event) => {
		const data = JSON.parse(event.data) as unknown;
		for (const h of handlers) h(data);
	};

	es.onerror = () => {};

	return {
		subscribe(handler: CommentHandler) {
			handlers.add(handler);
			return () => handlers.delete(handler);
		},
		close() {
			es.close();
		}
	};
}
