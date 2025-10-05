export function formatDate(date: Date) {
	return new Date(date).toLocaleString(undefined, {
		year: 'numeric',
		month: 'long',
		day: 'numeric',
		hour: 'numeric',
		minute: '2-digit',
		hour12: true
	});
}
