export function formatMinutes(minutes: number) {
	const units = [
		{ label: 'Year', minutes: 525600 },
		{ label: 'Month', minutes: 43800 },
		{ label: 'Day', minutes: 1440 },
		{ label: 'Hour', minutes: 60 },
		{ label: 'Minute', minutes: 1 }
	];

	const result: string[] = [];

	for (const unit of units) {
		const count = Math.floor(minutes / unit.minutes);
		if (count > 0) {
			result.push(`${count} ${unit.label}${count > 1 ? 's' : ''}`);
			minutes %= unit.minutes;
		}
	}

	return result.join(' and ');
}
