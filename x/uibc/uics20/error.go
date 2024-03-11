package uics20

type errMemoValidation struct {
	internal error
}

func (e errMemoValidation) Error() string {
	return "ics20 memo validation error " + e.internal.Error()
}

func (e errMemoValidation) Unwrap() error { return e.internal }
func (e errMemoValidation) Is(err error) bool {
	_, ok := err.(errMemoValidation)
	return ok
}
