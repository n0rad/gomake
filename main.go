package main

import "github.com/n0rad/gomake/src"

var commitHash string
var version string
var buildDate string

func main() {
	src.Main(commitHash, version, buildDate)
}