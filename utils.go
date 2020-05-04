package gomake

import (
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
	"time"
)

var t = true
var f = false
var True = &t
var False = &f

//func CommandDurationWrapper(f func(cmd *cobra.Command, args []string) error) func(*cobra.Command, []string) error {
//	return func(cmd *cobra.Command, args []string) error {
//		start := time.Now()
//		err := f(cmd, args)
//		diff := time.Now().Sub(start)
//		duration := diff.Round(time.Second).String()
//		ColorPrintln(cmd.Use+" duration : "+duration, HYellow)
//		return err
//	}
//}

func IsDirectoryEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func CommandDurationWrapper(cmd *cobra.Command, f func() error) error {
	start := time.Now()
	err := f()
	diff := time.Now().Sub(start)
	duration := diff.Round(time.Second)
	if duration > 0 {
		ColorPrintln(strings.Title(cmd.Use)+" : "+duration.String(), HBlue)
	}
	return err
}

func RegisterLogLevelParser(cmd *cobra.Command) {
	var logLevel string

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if logLevel != "" {
			level, err := logs.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			logs.SetLevel(level)
		}
		return nil
	}

	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "", "Set log level")

}
