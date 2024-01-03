package modelgen

type Context struct {
	State
}

type State struct {
	Require bool   // is current object required
	Path    string // current object's path
}
