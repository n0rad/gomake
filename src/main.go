package main

import (
	"fmt"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"os"
	"time"
)

const goMake = "gomake"

var commitHash string
var version string
var buildDate string

func main() {
	startTime := time.Now()

	argParser, err := prepareArgParser()
	if err != nil {
		logs.WithE(err).Fatal("Wrong arguments")
		os.Exit(1)
	}

	if err := argParser.Execute(); err != nil {
		logs.WithE(err).Fatal("Make failed")
		os.Exit(1)
	}
	logs.WithField("duration", time.Now().Sub(startTime)).Info("finished")
}

func displayVersionAndExit() {
	fmt.Println(goMake)
	if version == "" {
		version = "0"
	}
	fmt.Printf("Version    : %s\n", version)
	if buildDate != "" {
		fmt.Printf("Build date : %s\n", buildDate)
	}
	if commitHash != "" {
		fmt.Printf("CommitHash : %s\n", commitHash)
	}
	os.Exit(0)
}
