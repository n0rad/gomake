package gomake

import (
	"fmt"
	"github.com/n0rad/go-erlog/errs"
	"strings"
	"time"
)

func IsGitWorkTreeClean() error {
	if err := Exec("git", "update-index", "-q", "--ignore-submodules", "--refresh"); err != nil {
		return errs.WithE(err, "failed to update git index")
	}

	if err := Exec("git", "diff-files", "--quiet", "--ignore-submodules", "--"); err != nil {
		return errs.WithE(err, "You have unstaged changes")
	}

	if _, err := ExecGetStdout("git", "diff-files", "--name-status", "-r", "--ignore-submodules", "--"); err != nil {
		return errs.WithE(err, "You have uncommitted changes")
	}

	if err := Exec("git", "diff-index", "--cached", "--quiet", "HEAD", "--ignore-submodules", "--"); err != nil {
		return errs.WithE(err, "You have unstaged changes in the index")
	}

	if _, err := ExecGetStdout("git", "diff-index", "--cached", "--name-status", "-r", "--ignore-submodules", "HEAD", "--"); err != nil {
		return errs.WithE(err, "You have uncommitted changes in the index")
	}

	return nil
}

func GeneratedVersion() (string, error) {
	githash, err := ExecGetStdout("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", errs.WithE(err, "Failed to get git commit hash")
	}
	now := time.Now()
	return fmt.Sprintf("%s.%s.%s-%s",
		"1",
		now.Format("20060102"),
		strings.TrimLeft(now.Format("150405"), "0"),
		githash), nil

}
