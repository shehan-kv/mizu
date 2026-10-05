package billing

type CurrencyName string

func NewCurrencyName(name string) (CurrencyName, error) {
	if name == "" {
		return "", ErrBillingCurrencyNameCannotBeEmpty
	}

	return CurrencyName(name), nil
}

func (c CurrencyName) String() string {
	return string(c)
}
