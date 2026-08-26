export type ModalId = 'profile' | 'settings' | 'security' | 'logout' | 'checkout' | 'lightbox' | 'agegate' | 'login' | 'register' | 'forgot-password' | 'add-to-list' | 'new-message' | 'boost';

let active = $state<ModalId | null>(null);

export const modal = {
	get current() {
		return active;
	},
	open(id: ModalId) {
		active = id;
	},
	close(id?: ModalId) {
		if (!id || active === id) active = null;
	},
	closeAll() {
		active = null;
	},
	isOpen(id: ModalId) {
		return active === id;
	},
};
