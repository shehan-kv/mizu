package billing

type Currency struct {
	name          CurrencyName
	symbol        CurrencySymbol
	code          CurrencyCode
	decimalPlaces CurrencyDecimals
}

func NewCurrency(
	name CurrencyName,
	symbol CurrencySymbol,
	code CurrencyCode,
	decimalPlaces CurrencyDecimals,
) Currency {

	return Currency{
		name:          name,
		symbol:        symbol,
		code:          code,
		decimalPlaces: decimalPlaces,
	}
}

func (c Currency) Name() CurrencyName {
	return c.name
}

func (c Currency) Symbol() CurrencySymbol {
	return c.symbol
}

func (c Currency) Code() CurrencyCode {
	return c.code
}

func (c Currency) DecimalPlaces() CurrencyDecimals {
	return c.decimalPlaces
}
