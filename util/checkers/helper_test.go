package checkers

import (
	"strings"
)

func errsToStr(errs []error) string {
	strs := make([]string, len(errs))
	for i := range errs {
		strs[i] = errs[i].Error()
	}
	return strings.Join(strs, "    ")
}
