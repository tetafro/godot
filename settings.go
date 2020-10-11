package godot

// Settings contains linter settings.
type Settings struct {
	// Which comments to check (top level declarations, top level, all).
	Scope Scope

	// Check periods at the end of sentences.
	Period bool
}

// Scope sets which comments should be checked.
type Scope string

// List of available check scopes.
const (
	// DeclScope is for top level declaration comments.
	DeclScope Scope = "decl"
	// TopLevelScope is for all top level comments.
	TopLevelScope Scope = "top"
	// AllScope is for all comments.
	AllScope Scope = "all"
)
