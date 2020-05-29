package loyalty

// CreateUser command
type CreateUser struct {
	CommandModel
	Username string `json:"username"`
}

// DeleteUser command
type DeleteUser struct {
	CommandModel
	Username string `json:"username"`
}
