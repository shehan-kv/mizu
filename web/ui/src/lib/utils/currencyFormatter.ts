export function currencyFormatter(currency: string, amount: Intl.StringNumericLiteral | number) {
	return new Intl.NumberFormat(undefined, {
		style: 'currency',
		currencyDisplay: 'symbol',
		currency
	}).format(amount);
}
