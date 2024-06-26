package tsdk

import "math/rand"

func GenerateString(length uint) string {
	// character set used for generating a random string in GenerateString
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charset[rand.Intn(len(charset))] //nolint
	}
	return string(bytes)
}
