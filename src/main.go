package src

import (
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"os"
	"time"
)

const goMake = "gomake"

var CommitHash string
var Version string
var BuildDate string

func Main(commitHash string, version string, buildDate string) {
	startTime := time.Now()
	CommitHash = commitHash
	Version = version
	BuildDate = buildDate

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
