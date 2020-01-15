package annex

import "github.com/G-Node/git-module"

const (
	IdentStr = "/annex/objects"
)

func Upgrade(dir string) (string, error) {
	cmd := git.NewACommand("upgrade")
	return cmd.RunInDir(dir)
}
