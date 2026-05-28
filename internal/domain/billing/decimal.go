package billing

import "github.com/cockroachdb/apd/v3"

type Decimal struct {
	value *apd.Decimal
}

const decimalPrecision uint32 = 19

func NewDecimal(value string) (Decimal, error) {
	d, _, err := apd.NewFromString(value)
	if err != nil {
		return Decimal{}, ErrBillingInvalidDecimal
	}

	return Decimal{value: d}, nil
}

func (d Decimal) Add(other Decimal) (Decimal, error) {
	ctx := apd.BaseContext.WithPrecision(decimalPrecision)
	result := new(apd.Decimal)

	if _, err := ctx.Add(result, d.value, other.value); err != nil {
		return Decimal{}, ErrBillingAdditionFailed
	}

	return Decimal{value: result}, nil
}

func (d Decimal) Subtract(other Decimal) (Decimal, error) {
	ctx := apd.BaseContext.WithPrecision(decimalPrecision)
	result := new(apd.Decimal)

	if _, err := ctx.Sub(result, d.value, other.value); err != nil {
		return Decimal{}, ErrBillingSubtractionFailed
	}

	return Decimal{value: result}, nil
}

func (d Decimal) Multiply(other Decimal) (Decimal, error) {
	ctx := apd.BaseContext.WithPrecision(decimalPrecision)
	result := new(apd.Decimal)

	if _, err := ctx.Mul(result, d.value, other.value); err != nil {
		return Decimal{}, ErrBillingMultiplicationFailed
	}

	return Decimal{value: result}, nil
}

func (d Decimal) Divide(other Decimal) (Decimal, error) {
	ctx := apd.BaseContext.WithPrecision(decimalPrecision)
	result := new(apd.Decimal)

	if _, err := ctx.Quo(result, d.value, other.value); err != nil {
		return Decimal{}, ErrBillingDivisionFailed
	}

	return Decimal{value: result}, nil
}

func (d Decimal) GreaterThan(other Decimal) bool {
	return d.value.Cmp(other.value) == 1
}

func (d Decimal) Equals(other Decimal) bool {

	return d.value.Cmp(other.value) == 0
}

func (d Decimal) String() string {
	return d.value.String()
}
