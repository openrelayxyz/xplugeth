package xplugeth


// Test describes a go test package and lists the name of tests that can be used to verify a Patchset was applied successfully.
type Test struct {
	Package string `json:"package"`
	TestNames []string `json:"test"`
}

// Patchset describes a git remote and ref that can be cherry-picked onto Geth to enable a hook, along with Tests to verify that
// the patchset was applied correctly.
type Patchset struct {
	Remote string `json:"remote"`
	Ref string `json:"ref"`
	Tests []Test `json:"tests"`
}