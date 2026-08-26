export const addToListTarget = $state<{ comicId: string | null; title: string }>({ comicId: null, title: '' });

export function setAddToList(comicId: string, title: string) {
	addToListTarget.comicId = comicId;
	addToListTarget.title = title;
}

export function clearAddToList() {
	addToListTarget.comicId = null;
	addToListTarget.title = '';
}
