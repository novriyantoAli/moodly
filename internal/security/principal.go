package security

type Principal struct {
	UserID      uint
	Roles       []string
	Permissions []string
}