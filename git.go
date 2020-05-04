package gomake

import "github.com/n0rad/go-erlog/errs"

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
