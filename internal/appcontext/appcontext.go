package appcontext

// Context holds application-wide execution settings.
type Context struct {
	Debug  bool
	NoOp   bool
	Silent bool // Add Silent field
}