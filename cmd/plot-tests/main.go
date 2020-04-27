// testplot allows plotting `go test -json` output as a cascade to see
// testing parallelism and sequencing.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	//packageHeight := flag.Float64("package-height", 12, "height of a package line")
	//testHeight := flag.Float64("test-height", 3, "height of a test line")
	//collapseHeight := flag.Float64("collapse-height", 1, "height of a test line")
	//pxPerSecond := flag.Float64("pixels-per-second", 1, "how many pixels a single second is")
	// collapsePkgDuration := flag.Duration("ignore-package", 0*time.Second, "ignore packages with shorter duration")
	// collapseTestDuration := flag.Duration("ignore-test", 0*time.Second, "ignore tests with shorter duration")

	flag.Parse()

	//PackageHeight := *packageHeight
	//TestHeight := *testHeight
	//CollapseHeight := *collapseHeight
	//PxPerSecond := *pxPerSecond

	testsuite := NewTestSuite()

	var in io.Reader
	if fname := flag.Arg(0); fname != "" {
		var err error
		in, err = os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open %q: %w\n", fname, err)
			os.Exit(1)
		}
	} else {
		in = os.Stdin
	}

	dec := NewEventDecoder(bufio.NewReader(in))
	for {
		event, err := dec.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "failed to decode: %v\n", err)
			break
		}
		testsuite.AddEvent(event)
	}

	testsuite.Sort()

	rendered := RenderSVG(testsuite)

	os.Stdout.Write(rendered)
}
