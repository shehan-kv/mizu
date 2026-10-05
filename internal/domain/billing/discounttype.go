package billing

type DiscountType string

const (
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypePercentage DiscountType = "percentage"
)

func NewDiscountType(dType string) (DiscountType, error) {

	dt := DiscountType(dType)

	switch dt {
	case DiscountTypeFixed, DiscountTypePercentage:
		return dt, nil
	default:
		return "", ErrBillingInvalidDiscountType
	}
}

func (d DiscountType) String() string {
	return string(d)
}
