package fixtures

func validateUser(user User) error {
	if user.name == "" {
		return errUserName
	}
	if user.age < 0 {
		return errUserAge
	}
	if user.email == "" {
		return errUserEmail
	}
	user.normalized = true
	user.checked = true
	return nil
}
