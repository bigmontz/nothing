package repository

type userError struct {
	err error
}

func (u userError) IsUserError() bool {
	return true
}

func (u userError) Error() string {
	return u.err.Error()
}
