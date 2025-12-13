package gomake

import (
	"fmt"
	"strings"
	"time"

	"github.com/n0rad/go-erlog/errs"
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

func GeneratedVersionTime(now time.Time) (string, error) {
	githash, err := ExecGetStdout("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", errs.WithE(err, "Failed to get git commit hash")
	}
	hms := strings.TrimLeft(now.Format("150405"), "0")
	if hms == "" {
		hms = "0"
	}
	return fmt.Sprintf("%s.%s.%s-H%s",
		"1",
		now.Format("20060102"),
		hms,
		githash), nil
}

func GeneratedVersion() (string, error) {
	return GeneratedVersionTime(time.Now())
}
