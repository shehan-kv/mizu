export function toTitleCase(text: string): string {
	return text
		.toLowerCase()
		.split(/[\s-_]+/) // split on space, dash, or underscore
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(' ');
}
