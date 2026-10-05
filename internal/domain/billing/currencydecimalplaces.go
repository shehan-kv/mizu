package billing

type CurrencyDecimals int

func NewCurrencyDecimals(decimals int) (CurrencyDecimals, error) {
	if decimals < 0 {
		return 0, ErrBillingCurrencyDecimalsCannotBeNegative
	}

	return CurrencyDecimals(decimals), nil
}

func (c CurrencyDecimals) Int() int {
	return int(c)
}
