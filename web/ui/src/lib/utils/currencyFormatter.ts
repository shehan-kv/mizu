export function currencyFormatter(currency: string, amount: Intl.StringNumericLiteral | number) {
	return new Intl.NumberFormat('en-US', {
		style: 'currency',
		currency
	}).format(amount);
}
