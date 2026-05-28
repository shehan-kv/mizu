package billing

type CurrencySymbol string

func NewCurrencySymbol(symbol string) (CurrencySymbol, error) {
	if symbol == "" {
		return "", ErrBillingCurrencySymbolCannotBeEmpty
	}

	return CurrencySymbol(symbol), nil
}

func (c CurrencySymbol) String() string {
	return string(c)
}
