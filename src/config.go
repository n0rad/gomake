package src

type Config struct {
	name string
	repository string
	version string
	targetDirectory string
}

func newConfig() *Config {
	return &Config{
		targetDirectory: "/dist",
	}
}
