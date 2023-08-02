package errors

type LeverageError struct {
	Err           error
	IsRecoverable bool
}

func (le *LeverageError) Error() string {
	return le.Err.Error()
}

func Wrap(err error, isRecoverable bool) *LeverageError {
	return &LeverageError{
		Err:           err,
		IsRecoverable: isRecoverable,
	}
}
