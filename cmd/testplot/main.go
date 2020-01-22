// testplot allows plotting `go test -json` output as a cascade to see
// testing parallelism and sequencing.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"html/template"
	"image/color"
	"io"
	"math"
	"os"
	"sort"
	"time"

	"github.com/loov/diagram"
)

func main() {
	packageHeight := flag.Float64("package-height", 12, "height of a package line")
	testHeight := flag.Float64("test-height", 3, "height of a test line")
	pxPerSecond := flag.Float64("pixels-per-second", 1, "how many pixels a single second is")

	flag.Parse()

	PackageHeight := *packageHeight
	TestHeight := *testHeight
	PxPerSecond := *pxPerSecond

	testsuite := NewTestSuite()

	dec := json.NewDecoder(bufio.NewReader(os.Stdin))
	for {
		var event Event
		err := dec.Decode(&event)
		if err == io.EOF {
			break
		}
		if event.Time.IsZero() {
			continue
		}

		testsuite.Extend(event.Time)
		if event.Package == "" {
			continue
		}

		pkg := testsuite.Package(event.Package)
		pkg.Extend(event.Time)
		if event.Test == "" {
			continue
		}

		test := pkg.Test(event.Test)
		test.Extend(event.Time)
	}

	testsuite.Sort()

	ignorePkg := func(pkg *Package) bool {
		return pkg.Duration() < 5*time.Second
	}

	ignore := func(t *Test) bool {
		return t.Duration() < 5*time.Second
	}

	totalWidth := testsuite.Duration().Seconds()*PxPerSecond + 1

	totalHeight := 0.0
	for _, pkg := range testsuite.Packages {
		if ignorePkg(pkg) {
			continue
		}
		totalHeight += PackageHeight
		for _, test := range pkg.Tests {
			if ignore(test) {
				continue
			}
			totalHeight += TestHeight
		}
	}

	canvas := diagram.NewSVG(math.Ceil(totalWidth)+PackageHeight*10, math.Ceil(totalHeight))
	boxes := canvas.Layer(0)
	text := canvas.Layer(1)
	_ = text

	tox := func(t time.Time) float64 {
		return t.Sub(testsuite.Start).Seconds() * PxPerSecond
	}

	y := 0.0
	for _, pkg := range testsuite.Packages {
		if ignorePkg(pkg) {
			continue
		}

		r := diagram.R(tox(pkg.Start), y, tox(pkg.Finish), y+PackageHeight)
		y += PackageHeight
		boxes.Rect(r, &diagram.Style{
			Fill: color.Gray{0xB0},
			Hint: pkg.Name + " " + pkg.Duration().String(),
		})
		text.Text(pkg.Name, diagram.Point{
			X: r.Min.X + 5.0,
			Y: (r.Min.Y + r.Max.Y) / 2,
		}, &diagram.Style{
			Stroke: color.Black,
			Size:   r.Size().Y - 2,
			Origin: diagram.Point{X: -1, Y: 0},
		})

		for i, test := range pkg.Tests {
			if ignore(test) {
				continue
			}

			r := diagram.R(tox(test.Start), y, tox(test.Finish), y+TestHeight)
			y += TestHeight
			boxes.Rect(r, &diagram.Style{
				Fill: color.Gray{byte(0x40 * (1 + i%2))},
				Hint: test.Name + " " + test.Duration().String(),
			})
			//text.Text(test.Name, r.Min, &diagram.Style{
			//	Stroke: color.Black,
			//})
		}
	}

	os.Stdout.Write(canvas.Bytes())
}

type Span struct {
	Start  time.Time
	Finish time.Time
}

func (span *Span) Duration() time.Duration {
	return span.Finish.Sub(span.Start)
}

func (span *Span) Extend(t time.Time) {
	if span.Start.IsZero() {
		span.Start = t
	}
	if span.Finish.IsZero() {
		span.Finish = t
	}

	if t.Before(span.Start) {
		span.Start = t
	}
	if t.After(span.Finish) {
		span.Finish = t
	}
}

type TestSuite struct {
	Span

	ByName   map[string]*Package
	Packages []*Package
}

func NewTestSuite() *TestSuite {
	return &TestSuite{
		ByName: map[string]*Package{},
	}
}

func (ts *TestSuite) Package(pkgname string) *Package {
	if pkg, ok := ts.ByName[pkgname]; ok {
		return pkg
	}

	pkg := NewPackage(pkgname)
	ts.ByName[pkgname] = pkg
	ts.Packages = append(ts.Packages, pkg)

	return pkg
}

func (ts *TestSuite) Sort() {
	sort.Slice(ts.Packages, func(i, k int) bool {
		return ts.Packages[i].Start.Before(ts.Packages[k].Start)
	})

	for _, pkg := range ts.Packages {
		pkg.Sort()
	}
}

type Package struct {
	Name string
	Span
	ByName map[string]*Test
	Tests  []*Test
}

func (pkg *Package) Sort() {
	sort.Slice(pkg.Tests, func(i, k int) bool {
		return pkg.Tests[i].Start.Before(pkg.Tests[k].Start)
	})
}

func NewPackage(pkgname string) *Package {
	return &Package{
		Name:   pkgname,
		ByName: map[string]*Test{},
	}
}

func (pkg *Package) Test(testname string) *Test {
	if test, ok := pkg.ByName[testname]; ok {
		return test
	}

	test := NewTest(testname)
	pkg.ByName[testname] = test
	pkg.Tests = append(pkg.Tests, test)
	return test
}

type Test struct {
	Name string
	Span
}

func NewTest(testname string) *Test {
	return &Test{Name: testname}
}

type Event struct {
	Action  string
	Output  string
	Time    time.Time
	Package string
	Test    string
	Elapsed float64
}

var T = template.Must(template.New("").Parse(`
<!doctype html>
<html>
<head>
	<meta charset="utf-8">
	<style>
		.packages, .package, .span, .tests, .test {
			display: block;
			position: relative;
			white-space: nowrap;
		}

		.test {
			font-size: 0.7rem;
			background: #eee;
		}

		[data-duration<5] {
			display: none;
		}
	</style>
</head>
{{ $all := . }}
<body>
	<div class="packages">
		{{ range $pkg := .Packages }}
		<div class="package" data-duration="{{$pkg.Duration.Seconds}}">
			<div class="span" style="left:{{ ($pkg.Start.Sub $all.Start).Seconds }}px; width: {{ $pkg.Duration.Seconds }}px">{{ $pkg.Name }}</div>
			<div class="tests">
				{{ range $test := .Tests }}
				<div class="test" data-duration="{{$pkg.Duration.Seconds}}" style="left:{{ ($test.Start.Sub $all.Start).Seconds }}px; width: {{ $test.Duration.Seconds }}px">{{ $test.Name }}</div>
				{{ end }}
			</div>
		</div>
		{{ end }}
	</div>
</body>
</html>
`))
