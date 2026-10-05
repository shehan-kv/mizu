package billing

type Qty struct {
	value Decimal
}

func NewQty(qty string) (Qty, error) {
	dec, err := NewDecimal(qty)
	if err != nil {
		return Qty{}, err
	}
	zero, err := NewDecimal("0")
	if err != nil {
		return Qty{}, err
	}
	if !dec.GreaterThan(zero) {
		return Qty{}, ErrBillingQtyMustBePositive
	}
	return Qty{value: dec}, nil
}

func (q Qty) ToDecimal() Decimal {
	return q.value
}

func (q Qty) String() string {
	return q.value.String()
}
