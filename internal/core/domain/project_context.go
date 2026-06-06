package domain

type ProjectContext struct {
	Languages     []string
	Frameworks    []string
	BuildTools    []string
	TestCommands  []string
	DetectedFiles []string
}
