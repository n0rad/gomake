package gomake

import (
	"os"
	"strings"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
)

type StepRelease struct {
	project         *Project
	OsArchRelease   []string
	Upx             *bool
	Version         string
	Token           string
	GithubRelease   bool
	PostReleaseHook func(StepRelease) error // upload
}

func (c *StepRelease) Init(project *Project) error {
	c.project = project

	if len(c.OsArchRelease) == 0 {
		c.OsArchRelease = append(c.OsArchRelease, "linux-amd64")
		c.OsArchRelease = append(c.OsArchRelease, "darwin-amd64")
		c.OsArchRelease = append(c.OsArchRelease, "windows-amd64")
	}

	if c.Upx == nil {
		c.Upx = False
	}

	if c.Version == "" {
		version, err := c.project.versionFunc()
		if err != nil {
			return errs.WithE(err, "Failed to generate version")
		}
		c.Version = version
	}

	return nil
}

func (c *StepRelease) Name() string {
	return "release"
}

func (c *StepRelease) Project() *Project {
	return c.project
}

func (c *StepRelease) GetCommand() *cobra.Command {
	var token string
	var version string

	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "release",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := CommandDurationWrapper(cmd, func() error {
				ColorPrintln("Releasing", HGreen)

				c.Token = os.Getenv("GITHUB_TOKEN")
				if token != "" {
					c.Token = token
				}
				if version != "" {
					c.Version = version
				}

				// clean
				cleanCommand := c.project.steps["clean"].GetCommand()
				cleanCommand.SetArgs([]string{})
				if err := cleanCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, clean failed")
				}

				// build
				build := c.project.steps["build"].(*StepBuild)
				build.Upx = c.Upx
				build.Version = c.Version
				programs := build.Programs
				build.Programs = []Program{}

				for _, osArch := range c.OsArchRelease {
					for _, program := range programs {
						build.Programs = append(build.Programs, Program{
							BinaryName: program.BinaryName,
							Package:    program.Package,
							OsArch:     osArch,
						})
					}
				}
				if err := build.Init(c.project); err != nil {
					return errs.WithE(err, "Failed to re-init build for release")
				}
				buildCommand := build.GetCommand()
				buildCommand.SetArgs([]string{})
				if err := buildCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, build failed")
				}

				// test
				testCommand := c.project.steps["test"].GetCommand()
				testCommand.SetArgs([]string{})
				if err := testCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, test failed")
				}

				// check
				checkCommand := c.project.steps["check"].GetCommand()
				checkCommand.SetArgs([]string{})
				if err := checkCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, check failed")
				}

				if err := IsGitWorkTreeClean(); err != nil {
					return errs.WithE(err, "git repository is not clean")
				}

				// compressing
				for _, p := range build.Programs {
					if err := c.compressRelease(p); err != nil {
						return errs.WithE(err, "Compression failed")
					}
				}

				if c.PostReleaseHook != nil {
					if err := c.PostReleaseHook(*c); err != nil {
						return errs.WithE(err, "Post release hook failed")
					}
				}

				if c.GithubRelease {
					if err := c.releaseToGithub(); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return err
			}
			return c.project.processArgs(args)
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "token")
	cmd.Flags().StringVarP(&version, "version", "v", "", "version")
	RegisterLogLevelParser(cmd)

	return cmd
}

func (c StepRelease) releaseToGithub() error {
	gitRemoteUrl, err := ExecShellGetStdout(`git config --get remote.origin.url`)
	if err != nil {
		return errs.WithE(err, "Failed to get git remote origin url")
	}
	gitRemoteUrl = strings.TrimPrefix(gitRemoteUrl, "https://")
	gitRemoteUrl = strings.TrimPrefix(gitRemoteUrl, "git@")
	gitRemoteUrl = strings.TrimSuffix(gitRemoteUrl, ".git")
	gitRemoteUrl = strings.Replace(gitRemoteUrl, ":", "/", 1)

	if !strings.Contains(gitRemoteUrl, "github") {
		return errs.WithF(data.WithField("remoteUrl", gitRemoteUrl), "Remote is not a github url")
	}

	if c.Token == "" {
		return errs.With("github token is not set")
	}

	gitRemoteUrlSplit := strings.SplitN(gitRemoteUrl, "/", 2)
	if len(gitRemoteUrlSplit) < 2 {
		return errs.WithF(data.WithField("url", gitRemoteUrl), "Invalid github remote url")
	}
	githubRepoPath := gitRemoteUrlSplit[1]

	// detect default branch instead of hardcoding master
	defaultBranch, err := ExecShellGetStdout(`git symbolic-ref --short refs/remotes/origin/HEAD | sed 's@^origin/@@'`)
	if err != nil || defaultBranch == "" {
		// fallback to commonly used branches
		logs.WithError(err).Warn("Failed to detect default branch, falling back to master/main")
		defaultBranch = "master"
		if _, errMain := ExecShellGetStdout(`git show-ref --verify --quiet refs/heads/main && echo main || echo`); errMain == nil {
			// if main exists locally, prefer it
			candidate, _ := ExecShellGetStdout(`git rev-parse --abbrev-ref main || echo`)
			candidate = strings.TrimSpace(candidate)
			if candidate != "" {
				defaultBranch = candidate
			}
		}
	}
	defaultBranch = strings.TrimSpace(defaultBranch)

	posturl, err := ExecShellGetStdout(`curl -H "Authorization: token ` + c.Token + `" --data "{\"tag_name\": \"v` + c.Version + `\",\"target_commitish\": \"` + defaultBranch + `\",\"name\": \"v` + c.Version + `\",\"body\": \"Release of version ` + c.Version + `\",\"draft\": false,\"prerelease\": false}" https://api.github.com/repos/` + githubRepoPath + `/releases | grep "\"upload_url\"" | sed -ne 's/.*\(http[^"]*\).*/\1/p'`)
	if err != nil {
		return errs.WithE(err, "Failed to get github file post url")
	}
	posturl = strings.SplitN(posturl, "{", 2)[0]

	for _, osArch := range c.OsArchRelease {
		releaseFile := c.project.name + "-" + osArch + ".tar.gz"
		logs.WithField("file", releaseFile).Info("Uploading file")

		if err := Exec("curl", "-H", "Authorization: token "+c.Token, "-i", "-X", "POST", "-H", "Content-Type: application/x-gzip", "--data-binary", "@dist/"+releaseFile, posturl+"?name="+releaseFile+"&label="+releaseFile); err != nil {
			return errs.WithEF(err, data.WithField("file", releaseFile), "Failed to upload file")
		}
	}
	return nil
}

func (c StepRelease) compressRelease(p Program) error {
	if err := os.Chdir("./dist"); err != nil {
		return errs.WithE(err, "Failed to chdir to dist")
	}
	defer os.Chdir("../")
	fileToWrite, err := os.OpenFile(c.project.name+"-"+p.OsArch+".tar.gz", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errs.WithE(err, "Failed to open compressed release file") // TODO
	}
	defer fileToWrite.Close()
	if err := CompressToTarGzDirectory(c.project.name+"-"+p.OsArch, fileToWrite); err != nil {
		return errs.WithE(err, "Failed to compress dir to tar.gz")
	}
	return nil
}
