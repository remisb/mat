package conf

import "reflect"

// Args holds command line arguments after flags have been parsed.
type Args struct {
	Args       []string
	ConfigPath string
}

func NewConfigArgs(args []string) Args {
	return Args{
		Args: args,
	}
}

// argsT is used by Parse and Usage to detect struct fields of the Args type.
var argsT = reflect.TypeOf(Args{})

// Num returns the i'th argument in the Args slice. It returns an empty string
// the request element is not present.
func (a Args) Num(i int) string {
	if i < 0 || i >= len(a.Args) {
		return ""
	}
	return a.Args[i]
}
