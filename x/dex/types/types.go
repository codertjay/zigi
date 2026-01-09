package types

// Allowed pool formulas
var (
	FormulaConstantProduct = "constant_product"
	AllowedFormulas        = map[string]struct{}{
		FormulaConstantProduct: {},
	}
)

func IsValidFormula(formula string) bool {
	_, ok := AllowedFormulas[formula]
	return ok
}
