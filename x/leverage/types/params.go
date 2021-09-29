package types

import (
	"gopkg.in/yaml.v3"
)

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
