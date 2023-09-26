package checkers

func MergeErrors(errs1, errs2 []error) []error {
	if len(errs2) > 0 {
		return append(errs1, errs2...)
	}
	return errs1
}
