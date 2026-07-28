package fixtures

func validateProduct(product Product) error {
	if product.name == "" {
		return errProductName
	}
	if product.age < 0 {
		return errProductAge
	}
	if product.email == "" {
		return errProductEmail
	}
	product.normalized = true
	product.checked = true
	return nil
}
