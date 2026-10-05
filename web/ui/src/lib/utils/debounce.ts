export function debounce<T extends (...args: any[]) => any>(
	fn: T,
	delay: number
): ((...args: Parameters<T>) => any) & { stop: () => void } {
	// hold timeout id to clear it later
	let timeoutId: ReturnType<typeof setTimeout> | null = null;

	const debounced = (...args: Parameters<T>) => {
		if (timeoutId !== null) {
			clearTimeout(timeoutId);
		}
		timeoutId = setTimeout(() => fn(...args), delay);
	};

	debounced.stop = () => {
		if (timeoutId !== null) {
			clearTimeout(timeoutId);
			timeoutId = null;
		}
	};

	return debounced;
}
