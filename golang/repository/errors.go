package repository

type userError struct {
	err      error
	notFound bool
}

func (u userError) IsUserError() bool {
	return true
}

func (u userError) NotFound() bool {
	return u.notFound
}

func (u userError) Error() string {
	return u.err.Error()
}
