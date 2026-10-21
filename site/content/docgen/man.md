# Generating Man Pages For Your Own cobra.Command

Generating man pages from a cobra command is incredibly easy. An example is as follows:

```go
package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "my test program",
	}
	header := &doc.GenManHeader{
		Title: "MINE",
		Section: "3",
	}
	dir, err := os.MkdirTemp("", "my-program-man")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	err = doc.GenManTree(cmd, header, dir)
	if err != nil {
		log.Fatal(err)
	}
}
```

That will get you a man page `test.3` inside that directory
