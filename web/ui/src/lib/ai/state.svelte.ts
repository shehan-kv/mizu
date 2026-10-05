export type AIStatus =
	| 'disabled'
	| 'checking'
	| 'available'
	| 'downloading'
	| 'loading'
	| 'ready'
	| 'unavailable'
	| 'error';

export const aiState = $state({
	status: 'disabled' as AIStatus,
	enabled: false,
	progress: 0
});
