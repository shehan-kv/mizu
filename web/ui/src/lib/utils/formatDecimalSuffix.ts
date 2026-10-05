export function formatDecimalSuffix(num: Intl.StringNumericLiteral) {
	return Intl.NumberFormat(undefined, {
		maximumFractionDigits: 2,
		minimumFractionDigits: 2,
		notation: 'compact',
		compactDisplay: 'short'
	}).format(num);
}
