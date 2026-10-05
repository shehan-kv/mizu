package billing

type TaxType string

const (
	TaxTypeFixed      TaxType = "fixed"
	TaxTypePercentage TaxType = "percentage"
)

func NewTaxType(tType string) (TaxType, error) {

	t := TaxType(tType)

	switch t {
	case TaxTypeFixed, TaxTypePercentage:
		return t, nil
	default:
		return "", ErrBillingInvalidTaxType
	}
}

func (t TaxType) String() string {
	return string(t)
}
