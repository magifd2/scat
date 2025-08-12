package appcontext

// CtxKey is the key for the application context in a context.Context.
type CtxKeyType struct{}

var CtxKey = CtxKeyType{}

// Context holds application-wide execution settings.
type Context struct {
	Debug      bool
	NoOp       bool
	Silent     bool // Add Silent field
	ConfigPath string // Path to the config file, if specified
}

// NewContext creates a new application context.
func NewContext(debug, noOp, silent bool, configPath string) Context {
	return Context{
		Debug:      debug,
		NoOp:       noOp,
		Silent:     silent,
		ConfigPath: configPath,
	}
}
