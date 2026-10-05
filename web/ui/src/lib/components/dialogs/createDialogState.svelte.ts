export function createDialogState() {
	let state = $state({ isOpen: false });

	return {
		get isOpen() {
			return state.isOpen;
		},
		set isOpen(value) {
			state.isOpen = value;
		},
		open() {
			state.isOpen = true;
		},
		close() {
			state.isOpen = false;
		}
	};
}
