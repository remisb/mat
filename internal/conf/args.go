package conf

// Args holds command line arguments after flags have been parsed.
type Args struct {
	Args       []string
	ConfigPath string
}

// NewConfigArgs is a factory function which create Args structure. command line args should
// be passed as an argument.
func NewConfigArgs(args []string) Args {
	return Args{
		Args: args,
	}
}

// Num returns the i'th argument in the Args slice. It returns an empty string
// the request element is not present.
func (a Args) Num(i int) string {
	if i < 0 || i >= len(a.Args) {
		return ""
	}
	return a.Args[i]
}
