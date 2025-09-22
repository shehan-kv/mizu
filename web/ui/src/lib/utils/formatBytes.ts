export function formatBytes(bytes: number, decimals = 2) {
	if (bytes === 0) return '0 B';
	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

	const i = Math.floor(Math.log(bytes) / Math.log(k));
	const value = bytes / Math.pow(k, i);

	return `${value.toFixed(decimals).replace(/\.?0+$/, '')} ${sizes[i]}`;
}
