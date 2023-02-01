package util

// Panic panics on error
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}
