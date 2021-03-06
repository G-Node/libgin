package annex

import (
	"fmt"

	"github.com/gogs/git-module"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

func Init(dir string, args ...string) ([]byte, error) {
	cmd := git.NewCommand("annex", "init")
	return cmd.AddArgs(args...).RunInDir(dir)
}

func Uninit(dir string, args ...string) ([]byte, error) {
	cmd := git.NewCommand("annex", "uninit")
	return cmd.AddArgs(args...).RunInDir(dir)
}

func Worm(dir string) ([]byte, error) {
	cmd := git.NewCommand("config", "annex.backends", "WORM")
	return cmd.RunInDir(dir)
}

func MD5(dir string) ([]byte, error) {
	cmd := git.NewCommand("config", "annex.backends", "MD5")
	return cmd.RunInDir(dir)
}

func Sync(dir string, args ...string) ([]byte, error) {
	cmd := git.NewCommand("annex", "sync")
	return cmd.AddArgs(args...).RunInDir(dir)
}

func Add(dir string, args ...string) ([]byte, error) {
	cmd := git.NewCommand("annex", "add")
	cmd.AddArgs(args...)
	return cmd.RunInDir(dir)
}

func SetAddUnlocked(dir string) ([]byte, error) {
	cmd := git.NewCommand("config", "annex.addunlocked", "true")
	return cmd.RunInDir(dir)
}

func SetAnnexSizeFilter(dir string, size int64) ([]byte, error) {
	cmd := git.NewCommand("config", "annex.largefiles", fmt.Sprintf("largerthan=%d", size))
	return cmd.RunInDir(dir)
}
