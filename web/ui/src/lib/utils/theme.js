export function toggleTheme() {
	if (
		localStorage.getItem('theme') === 'dark' ||
		(!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches)
	) {
		document.documentElement.classList.remove('dark');
		localStorage.setItem('theme', 'light');
	} else {
		document.documentElement.classList.add('dark');
		localStorage.setItem('theme', 'dark');
	}
}
