# Gomake

go project builder tool written in go


## Usage 

Create for example a `hack/gomake.go` file in your project:

```go
package main

import "github.com/n0rad/gomake"

func main() {
	gomake.ProjectBuilder().
		WithStep(&gomake.StepBuild{
			BinaryName: "hdm",
		}).
		MustBuild().MustExecute()
}
```

In root directory simplify calling gomake with a `Makefile` or a shell `script`

Makefile:
```makefile
.DEFAULT_GOAL := all
GOMAKE_PATH := ./hack

all:
	go run $(GOMAKE_PATH)

clean:
	go run $(GOMAKE_PATH) clean

build:
	go run $(GOMAKE_PATH) build

test:
	go run $(GOMAKE_PATH) test

quality:
	go run $(GOMAKE_PATH) quality

release:
	go run $(GOMAKE_PATH) release
```

shell script:
```shell script
#!/bin/sh
exec go run "$( cd "$(dirname "$0")" ; pwd -P )/hack" $@
``` 
