package billing

import "strings"

type CurrencyCode string

func NewCurrencyCode(code string) (CurrencyCode, error) {
	if code == "" {
		return "", ErrBillingCurrencyCodeCannotBeEmpty
	}
	if len(code) != 3 {
		return "", ErrBillingInvalidCurrencyCode
	}
	return CurrencyCode(strings.ToUpper(code)), nil
}

func (c CurrencyCode) String() string {
	return string(c)
}
