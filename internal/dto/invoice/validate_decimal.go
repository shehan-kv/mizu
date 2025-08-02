package invoice

import "github.com/cockroachdb/apd/v3"

func validateDecimal(d *apd.Decimal, precision, scale int32) bool {

	totalDigits := int32(len(d.Coeff.String()))

	exp := d.Exponent

	var beforeDecimal, afterDecimal int32

	if exp >= 0 {
		beforeDecimal = totalDigits + exp
		afterDecimal = 0
	} else {

		afterDecimal = -exp

		if afterDecimal >= totalDigits {
			beforeDecimal = 1
		} else {
			beforeDecimal = totalDigits - afterDecimal
		}
	}

	return beforeDecimal+afterDecimal <= precision && afterDecimal <= scale
}
