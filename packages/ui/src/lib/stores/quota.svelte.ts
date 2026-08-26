let tick = $state(0);

export const quotaRefresh = {
	get tick() {
		return tick;
	},
	bump() {
		tick++;
	},
};
