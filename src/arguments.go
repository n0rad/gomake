package src

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
	"fmt"
)

var workPath string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install binary to $GOPATH",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := newProject(workPath)
		if err != nil {
			logs.WithE(err).Fatal("Install command failed")
		}
		err = project.Install()
		if err != nil {
			logs.WithE(err).Fatal("Install command failed")
		}
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := newProject(workPath)
		if err != nil {
			logs.WithE(err).Fatal("Clean command failed")
		}
		err = project.Clean()
		if err != nil {
			logs.WithE(err).Fatal("Clean command failed")
		}
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := newProject(workPath)
		if err != nil {
			logs.WithE(err).Fatal("Build command failed")
		}
		err = project.internalCommand("build", args)
		if err != nil {
			logs.WithE(err).Fatal("Build command failed")
		}
	},
}

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "quality",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := newProject(workPath)
		if err != nil {
			logs.WithE(err).Fatal("Quality command failed")
		}
		err = project.internalCommand("quality", args)
		if err != nil {
			logs.WithE(err).Fatal("Quality command failed")
		}
	},
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "release",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := newProject(workPath)
		if err != nil {
			logs.WithE(err).Fatal("Release command failed")
		}
		err = project.internalCommand("release", args)
		if err != nil {
			logs.WithE(err).Fatal("Release command failed")
		}
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := newProject(workPath)
		if err != nil {
			logs.WithE(err).Fatal("Test command failed")
		}
		err = project.internalCommand("test", args)
		if err != nil {
			logs.WithE(err).Fatal("Test command failed")
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "gomake version",
	Long:  `Print gomake version info`,
	Run: func(cmd *cobra.Command, args []string) {
		displayVersionAndExit()
	},
}

func prepareArgParser() (*cobra.Command, error) {
	var err error
	var version bool
	var logLevel string
	var __ string

	var rootCmd = &cobra.Command{
		Use: "gomake",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if version {
				displayVersionAndExit()
			}

		},
	}

	rootCmd.PersistentFlags().StringVarP(&__, "log-level", "L", "info", "Set log level")
	logLevel, err = discoverStringArgument("L", "log-level", "info")
	if err != nil {
		return nil, err
	}

	level, err := logs.ParseLevel(logLevel)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("input", logLevel), "Cannot set log level")
	}
	logs.SetLevel(level)

	rootCmd.PersistentFlags().StringVarP(&__, "work-path", "W", ".", "Set the work path")
	workPath, err = discoverStringArgument("W", "work-path", ".")
	if err != nil {
		return nil, err
	}

	rootCmd.PersistentFlags().BoolVarP(&version, "version", "V", false, "Display dgr version")

	scripts := []string{}
	if files, err := ioutil.ReadDir(workPath + "/scripts"); err == nil {
		logs.WithField("path", workPath+"/scripts").Debug("Found scripts directory")
		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), "command-") {
				scriptFullPath := workPath + "/scripts/" + file.Name()
				files2 := strings.Split(file.Name()[len("command-"):], ".")
				cmd := &cobra.Command{
					Use:   files2[0],
					Short: "Run command from " + scriptFullPath,
					Run: func(cmd *cobra.Command, args []string) {
						if err := common.ExecCmd(scriptFullPath, args...); err != nil {
							logs.WithEF(err, data.WithField("script", scriptFullPath)).Fatal("External command failed")
						}
					},
				}
				rootCmd.AddCommand(cmd)
				scripts = append(scripts, files2[0])
			}
		}
	}

	rootCmd.AddCommand(versionCmd, cleanCmd, installCmd)
	for _, cmd := range []*cobra.Command{buildCmd, qualityCmd, releaseCmd, testCmd} {
		found := false
		for _, script := range scripts {
			if script == cmd.Use {
				found = true
				break
			}
		}
		if !found {
			rootCmd.AddCommand(cmd)
		}
	}

	return rootCmd, nil
}


func displayVersionAndExit() {
	fmt.Println(goMake)
	if Version == "" {
		Version = "0"
	}
	fmt.Printf("Version    : %s\n", Version)
	if BuildDate != "" {
		fmt.Printf("Build date : %s\n", BuildDate)
	}
	if CommitHash != "" {
		fmt.Printf("CommitHash : %s\n", CommitHash)
	}
	os.Exit(0)
}


func discoverStringArgument(shortName string, longName string, defaultValue string) (string, error) {
	workPathArgument := "--" + longName
	workPathArgumentAttached := workPathArgument + "="
	shortNameArgument := "-" + shortName
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--" {
			return defaultValue, nil
		} else if os.Args[i] == shortNameArgument || os.Args[i] == workPathArgument {
			if len(os.Args) <= i+1 {
				return defaultValue, errs.With("Missing --" + longName + " (-" + shortName + ") value")
			}
			return os.Args[i+1], nil
		} else if strings.HasPrefix(os.Args[i], workPathArgumentAttached) {
			return os.Args[i][len(workPathArgumentAttached):], nil
		}
	}
	return defaultValue, nil
}
