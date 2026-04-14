package gomake

import (
	"fmt"
	"strings"
	"time"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

func IsGitWorkTreeClean() error {
	if err := Exec("git", "update-index", "-q", "--ignore-submodules", "--refresh"); err != nil {
		return errs.WithE(err, "failed to update git index")
	}

	if unstagedFiles, err := ExecGetStdout("git", "diff-files", "--name-only", "--ignore-submodules", "--"); err != nil {
		return errs.WithE(err, "failed to check unstaged changes")
	} else if strings.TrimSpace(unstagedFiles) != "" {
		return errs.WithF(data.WithField("files", strings.TrimSpace(unstagedFiles)), "you have unstaged changes")
	}

	if stagedFiles, err := ExecGetStdout("git", "diff-index", "--cached", "--name-only", "--ignore-submodules", "HEAD", "--"); err != nil {
		return errs.WithE(err, "failed to check staged changes")
	} else if strings.TrimSpace(stagedFiles) != "" {
		return errs.WithF(data.WithField("files", strings.TrimSpace(stagedFiles)), "you have staged but uncommitted changes")
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
