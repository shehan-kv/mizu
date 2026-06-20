package billing

type Item struct {
	description  string
	qty          Qty
	unitPrice    Decimal
	discountRate Decimal
	discountType DiscountType
	taxRate      Decimal
	taxType      TaxType

	// precomputed values
	discountAmountPerUnit Decimal
	taxableBasePerUnit    Decimal
	taxAmountPerUnit      Decimal
	lineGross             Decimal
	lineDiscount          Decimal
	lineNet               Decimal
	lineTax               Decimal
	lineTotal             Decimal
}

func NewItem(
	description string,
	qty Qty,
	unitPrice Decimal,
	discountRate Decimal,
	discountType DiscountType,
	taxRate Decimal,
	taxType TaxType,
) (Item, error) {

	if description == "" {
		return Item{}, ErrBillingDescriptionCannotBeEmpty
	}

	discountAmtUnit, err := calculateDiscountAmountPerUnit(unitPrice, discountRate, discountType)
	if err != nil {
		return Item{}, err
	}

	qtyDecimal := qty.ToDecimal()

	taxableBasePerUnit, err := unitPrice.Subtract(discountAmtUnit)
	if err != nil {
		return Item{}, err
	}

	taxAmtUnit, err := calculateTaxAmountPerUnit(taxableBasePerUnit, taxRate, taxType)
	if err != nil {
		return Item{}, err
	}

	lineDiscount, err := discountAmtUnit.Multiply(qtyDecimal)
	if err != nil {
		return Item{}, err
	}

	lineGross, err := unitPrice.Multiply(qtyDecimal)
	if err != nil {
		return Item{}, err
	}

	lineNet, err := lineGross.Subtract(lineDiscount)
	if err != nil {
		return Item{}, err
	}

	lineTax, err := taxAmtUnit.Multiply(qtyDecimal)
	if err != nil {
		return Item{}, err
	}

	lineTotal, err := lineNet.Add(lineTax)
	if err != nil {
		return Item{}, err
	}

	return Item{
		description:  description,
		qty:          qty,
		unitPrice:    unitPrice,
		discountRate: discountRate,
		discountType: discountType,
		taxRate:      taxRate,
		taxType:      taxType,

		discountAmountPerUnit: discountAmtUnit,
		taxableBasePerUnit:    taxableBasePerUnit,
		taxAmountPerUnit:      taxAmtUnit,
		lineDiscount:          lineDiscount,
		lineTax:               lineTax,
		lineGross:             lineGross,
		lineNet:               lineNet,
		lineTotal:             lineTotal,
	}, nil
}

func RestoreItem(
	description string,
	qty Qty,
	unitPrice Decimal,
	discountRate Decimal,
	discountType DiscountType,
	taxRate Decimal,
	taxType TaxType,
	discountAmountPerUnit Decimal,
	taxableBasePerUnit Decimal,
	taxAmountPerUnit Decimal,
	lineDiscount Decimal,
	lineTax Decimal,
	lineGross Decimal,
	lineNet Decimal,
	lineTotal Decimal,
) Item {

	return Item{
		description:  description,
		qty:          qty,
		unitPrice:    unitPrice,
		discountRate: discountRate,
		discountType: discountType,
		taxRate:      taxRate,
		taxType:      taxType,

		discountAmountPerUnit: discountAmountPerUnit,
		taxableBasePerUnit:    taxableBasePerUnit,
		taxAmountPerUnit:      taxAmountPerUnit,
		lineDiscount:          lineDiscount,
		lineTax:               lineTax,
		lineGross:             lineGross,
		lineNet:               lineNet,
		lineTotal:             lineTotal,
	}
}

func (t Item) Description() string {
	return t.description
}

func (t Item) Qty() Qty {
	return t.qty
}

func (t Item) UnitPrice() Decimal {
	return t.unitPrice
}

func (t Item) DiscountRate() Decimal {
	return t.discountRate
}

func (t Item) DiscountType() DiscountType {
	return t.discountType
}

func (t Item) TaxRate() Decimal {
	return t.taxRate
}

func (t Item) TaxType() TaxType {
	return t.taxType
}

func (t Item) DiscountAmountPerUnit() Decimal {
	return t.discountAmountPerUnit
}

func (t Item) TaxAmountPerUnit() Decimal {
	return t.taxAmountPerUnit
}

func (t Item) TaxableBasePerUnit() Decimal {
	return t.taxableBasePerUnit
}

func (t Item) LineDiscount() Decimal {
	return t.lineDiscount
}

func (t Item) LineTax() Decimal {
	return t.lineTax
}

func (t Item) LineGross() Decimal {
	return t.lineGross
}

func (t Item) LineNet() Decimal {
	return t.lineNet
}

func (t Item) LineTotal() Decimal {
	return t.lineTotal
}

func (t Item) Equals(other Item) bool {
	return t == other
}

func percentageOf(rate Decimal, base Decimal) (Decimal, error) {
	hundred, err := NewDecimal("100")
	if err != nil {
		return Decimal{}, err
	}
	fraction, err := rate.Divide(hundred)
	if err != nil {
		return Decimal{}, err
	}
	return base.Multiply(fraction)
}

func calculateDiscountAmountPerUnit(unitPrice Decimal, discountRate Decimal, discountType DiscountType) (Decimal, error) {
	if discountType == DiscountTypePercentage {
		return percentageOf(discountRate, unitPrice)
	}
	return discountRate, nil
}

func calculateTaxAmountPerUnit(unitPrice Decimal, taxRate Decimal, taxType TaxType) (Decimal, error) {
	if taxType == TaxTypePercentage {
		return percentageOf(taxRate, unitPrice)
	}
	return taxRate, nil
}
