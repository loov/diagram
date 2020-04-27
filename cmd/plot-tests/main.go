// testplot allows plotting `go test -json` output as a cascade to see
// testing parallelism and sequencing.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	config := Config{
		PackageHeight: 20,
		TestHeight:    10,
		PxPerSecond:   2,

		IgnorePackage: 2 * time.Second,
		IgnoreTest:    2 * time.Second,
	}

	flag.Float64Var(&config.PackageHeight, "plot.package-height", config.PackageHeight, "height of a package span")
	flag.Float64Var(&config.TestHeight, "plot.test-height", config.TestHeight, "height of a test span")
	flag.Float64Var(&config.PxPerSecond, "plot.px-per-second", config.PxPerSecond, "how many pixels per second")

	flag.DurationVar(&config.IgnorePackage, "ignore-package", config.IgnorePackage, "ignore packages with shorter duration")
	flag.DurationVar(&config.IgnoreTest, "ignore-test", config.IgnoreTest, "ignore tests with shorter duration")

	flag.Parse()

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

	rendered := RenderSVG(config, testsuite)

	os.Stdout.Write(rendered)
}
