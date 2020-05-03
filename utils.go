package gomake

import (
	"github.com/spf13/cobra"
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
