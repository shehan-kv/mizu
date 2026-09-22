package billing

import "testing"

func newTestItem(
	t *testing.T,
	description string,
	qty string,
	unitPrice string,
	discountRate string,
	discountType DiscountType,
	taxRate string,
	taxType TaxType,
) Item {
	t.Helper()

	q, err := NewQty(qty)
	if err != nil {
		t.Fatal(err)
	}

	price, err := NewDecimal(unitPrice)
	if err != nil {
		t.Fatal(err)
	}

	discount, err := NewDecimal(discountRate)
	if err != nil {
		t.Fatal(err)
	}

	tax, err := NewDecimal(taxRate)
	if err != nil {
		t.Fatal(err)
	}

	item, err := NewItem(
		description,
		q,
		price,
		discount,
		discountType,
		tax,
		taxType,
	)
	if err != nil {
		t.Fatal(err)
	}

	return item
}

func TestNewItemWithFixedDiscountAndTax(t *testing.T) {
	item := newTestItem(
		t,
		"Development",
		"2",
		"100",
		"10",
		DiscountTypeFixed,
		"5",
		TaxTypeFixed,
	)

	tests := []struct {
		name string
		got  Decimal
		want string
	}{
		{
			name: "discount amount per unit",
			got:  item.DiscountAmountPerUnit(),
			want: "10",
		},
		{
			name: "taxable base per unit",
			got:  item.TaxableBasePerUnit(),
			want: "90",
		},
		{
			name: "tax amount per unit",
			got:  item.TaxAmountPerUnit(),
			want: "5",
		},
		{
			name: "line gross",
			got:  item.LineGross(),
			want: "200",
		},
		{
			name: "line discount",
			got:  item.LineDiscount(),
			want: "20",
		},
		{
			name: "line net",
			got:  item.LineNet(),
			want: "180",
		},
		{
			name: "line tax",
			got:  item.LineTax(),
			want: "10",
		},
		{
			name: "line total",
			got:  item.LineTotal(),
			want: "190",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.String() != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, tt.got.String())
			}
		})
	}

	if item.Description() != "Development" {
		t.Fatalf("expected description Development, got %q", item.Description())
	}

	expectedQty := mustDecimal(t, "2")
	if !item.Qty().ToDecimal().Equals(expectedQty) {
		t.Fatalf("expected quantity 2, got %s", item.Qty().String())
	}

	expectedUnitPrice := mustDecimal(t, "100")
	if !item.UnitPrice().Equals(expectedUnitPrice) {
		t.Fatalf("expected unit price 100, got %s", item.UnitPrice().String())
	}

	expectedDiscount := mustDecimal(t, "10")
	if !item.DiscountRate().Equals(expectedDiscount) {
		t.Fatalf("expected discount rate 10, got %s", item.DiscountRate().String())
	}

	if item.DiscountType() != DiscountTypeFixed {
		t.Fatalf("expected fixed discount type, got %s", item.DiscountType())
	}

	expectedTax := mustDecimal(t, "5")
	if !item.TaxRate().Equals(expectedTax) {
		t.Fatalf("expected tax rate 5, got %s", item.TaxRate().String())
	}

	if item.TaxType() != TaxTypeFixed {
		t.Fatalf("expected fixed tax type, got %s", item.TaxType())
	}
}

func TestNewItemWithPercentageDiscountAndTax(t *testing.T) {
	item := newTestItem(
		t,
		"Consulting",
		"2",
		"100",
		"10",
		DiscountTypePercentage,
		"20",
		TaxTypePercentage,
	)

	tests := []struct {
		name string
		got  Decimal
		want string
	}{
		{
			name: "discount amount per unit",
			got:  item.DiscountAmountPerUnit(),
			want: "10.00",
		},
		{
			name: "taxable base per unit",
			got:  item.TaxableBasePerUnit(),
			want: "90.00",
		},
		{
			name: "tax amount per unit",
			got:  item.TaxAmountPerUnit(),
			want: "18.0000",
		},
		{
			name: "line gross",
			got:  item.LineGross(),
			want: "200",
		},
		{
			name: "line discount",
			got:  item.LineDiscount(),
			want: "20.00",
		},
		{
			name: "line net",
			got:  item.LineNet(),
			want: "180.00",
		},
		{
			name: "line tax",
			got:  item.LineTax(),
			want: "36.0000",
		},
		{
			name: "line total",
			got:  item.LineTotal(),
			want: "216.0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedWant := mustDecimal(t, tt.want)
			if !tt.got.Equals(expectedWant) {
				t.Fatalf("expected %s, got %s", tt.want, tt.got.String())
			}
		})
	}
}

func TestNewItemRejectsEmptyDescription(t *testing.T) {
	qty, err := NewQty("1")
	if err != nil {
		t.Fatal(err)
	}

	price := mustDecimal(t, "100")
	discount := mustDecimal(t, "0")
	tax := mustDecimal(t, "0")

	_, err = NewItem(
		"",
		qty,
		price,
		discount,
		DiscountTypeFixed,
		tax,
		TaxTypeFixed,
	)

	if err != ErrBillingDescriptionCannotBeEmpty {
		t.Fatalf("expected %v, got %v", ErrBillingDescriptionCannotBeEmpty, err)
	}
}

func TestRestoreItem(t *testing.T) {
	qty, err := NewQty("2")
	if err != nil {
		t.Fatal(err)
	}

	unitPrice := mustDecimal(t, "100")
	discountRate := mustDecimal(t, "10")
	taxRate := mustDecimal(t, "5")

	discountAmountPerUnit := mustDecimal(t, "10")
	taxableBasePerUnit := mustDecimal(t, "90")
	taxAmountPerUnit := mustDecimal(t, "5")
	lineDiscount := mustDecimal(t, "20")
	lineTax := mustDecimal(t, "10")
	lineGross := mustDecimal(t, "200")
	lineNet := mustDecimal(t, "180")
	lineTotal := mustDecimal(t, "190")

	item := RestoreItem(
		"Development",
		qty,
		unitPrice,
		discountRate,
		DiscountTypeFixed,
		taxRate,
		TaxTypeFixed,
		discountAmountPerUnit,
		taxableBasePerUnit,
		taxAmountPerUnit,
		lineDiscount,
		lineTax,
		lineGross,
		lineNet,
		lineTotal,
	)

	if item.Description() != "Development" {
		t.Fatalf("expected description Development, got %q", item.Description())
	}

	expectedUnitDiscount := mustDecimal(t, "10")
	if !item.DiscountAmountPerUnit().Equals(expectedUnitDiscount) {
		t.Fatalf("expected discount amount 10, got %s", item.DiscountAmountPerUnit())
	}

	expectedUnitTaxBase := mustDecimal(t, "90")
	if !item.TaxableBasePerUnit().Equals(expectedUnitTaxBase) {
		t.Fatalf("expected taxable base 90, got %s", item.TaxableBasePerUnit())
	}

	expectedUnitTax := mustDecimal(t, "5")
	if !item.TaxAmountPerUnit().Equals(expectedUnitTax) {
		t.Fatalf("expected tax amount 5, got %s", item.TaxAmountPerUnit())
	}

	expectedLineDiscount := mustDecimal(t, "20")
	if !item.LineDiscount().Equals(expectedLineDiscount) {
		t.Fatalf("expected line discount 20, got %s", item.LineDiscount())
	}

	expectedLineTax := mustDecimal(t, "10")
	if !item.LineTax().Equals(expectedLineTax) {
		t.Fatalf("expected line tax 10, got %s", item.LineTax())
	}

	expectedLineGross := mustDecimal(t, "200")
	if !item.LineGross().Equals(expectedLineGross) {
		t.Fatalf("expected line gross 200, got %s", item.LineGross())
	}

	expectedLineNet := mustDecimal(t, "180")
	if !item.LineNet().Equals(expectedLineNet) {
		t.Fatalf("expected line net 180, got %s", item.LineNet())
	}

	expectedLineTotal := mustDecimal(t, "190")
	if !item.LineTotal().Equals(expectedLineTotal) {
		t.Fatalf("expected line total 190, got %s", item.LineTotal())
	}
}

func TestItemEquals(t *testing.T) {
	item1 := newTestItem(
		t,
		"Development",
		"2",
		"100",
		"10",
		DiscountTypeFixed,
		"5",
		TaxTypeFixed,
	)

	item2 := newTestItem(
		t,
		"Development",
		"2",
		"100",
		"10",
		DiscountTypeFixed,
		"5",
		TaxTypeFixed,
	)

	item3 := newTestItem(
		t,
		"Different",
		"2",
		"100",
		"10",
		DiscountTypeFixed,
		"5",
		TaxTypeFixed,
	)

	if !item1.Equals(item2) {
		t.Fatal("expected identical items to be equal")
	}

	if item1.Equals(item3) {
		t.Fatal("expected different items not to be equal")
	}
}
