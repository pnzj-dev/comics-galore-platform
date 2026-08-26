export const newMessageTarget = $state<{ userId: string | null; name: string }>({ userId: null, name: '' });

export function setNewMessage(userId: string, name: string) {
	newMessageTarget.userId = userId;
	newMessageTarget.name = name;
}

export function clearNewMessage() {
	newMessageTarget.userId = null;
	newMessageTarget.name = '';
}
