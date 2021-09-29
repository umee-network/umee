package types

import (
	"gopkg.in/yaml.v3"
)

// String implements the Stringer interface.
func (p UpdateAssetsProposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
