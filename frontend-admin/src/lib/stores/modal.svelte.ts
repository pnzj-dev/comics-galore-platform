export type ModalId = 'logout' | 'wizard';

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
