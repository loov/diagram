package main

import (
	"image/color"
	"time"

	"loov.dev/diagram"
)

type Config struct {
	PackageHeight float64
	TestHeight    float64
	PxPerSecond   float64

	IgnorePackage time.Duration
	IgnoreTest    time.Duration
}

type Plot struct {
	Config

	maxX float64
	y    float64
	line int
	span Span

	SVG   *diagram.SVG
	Grid  diagram.Canvas
	Spans diagram.Canvas
	Assoc diagram.Canvas
	Text  diagram.Canvas
}

func RenderSVG(ts *TestSuite) []byte {
	canvas := diagram.NewSVG(0, 0)

	plot := &Plot{
		Config: Config{
			PackageHeight: 20,
			TestHeight:    10,
			PxPerSecond:   2,

			IgnorePackage: 2 * time.Second,
			IgnoreTest:    2 * time.Second,
		},

		maxX: 0,
		y:    20,
		span: ts.Span,

		Grid:  canvas.Layer(0),
		Spans: canvas.Layer(1),
		Assoc: canvas.Layer(2),
		Text:  canvas.Layer(3),
	}

	for _, sub := range ts.Sub {
		plot.addPackage(sub)
	}

	plot.addGrid()
	width, height := plot.tox(ts.Finish)+150, plot.y
	canvas.Style = ""
	canvas.SetSize(width, height)

	canvas.Layer(-1).Rect(diagram.R(
		0, 0,
		width, height,
	), &diagram.Style{
		Fill: color.Gray{0xFF},
	})

	return canvas.Bytes()
}

func (p *Plot) tox(t time.Time) float64 {
	return 50 + t.Sub(p.span.Start).Seconds()*p.PxPerSecond
}

func (p *Plot) addGrid() {
	duration := p.span.Duration()

	minorTick := 10 * time.Second
	k := 1
	for tick := 0 * time.Second; tick < duration; tick += minorTick {
		k++
		if k%2 == 0 {
			continue
		}

		t := p.span.Start.Add(tick)
		p.Grid.Rect(diagram.R(
			p.tox(t), 0,
			p.tox(t.Add(minorTick)), p.y,
		), &diagram.Style{
			Fill: color.Gray{0xF8},
		})
	}

	majorTick := time.Minute
	for tick := 0 * time.Second; tick < duration; tick += majorTick {
		t := p.span.Start.Add(tick)

		p.Grid.Poly([]diagram.Point{
			{X: p.tox(t), Y: 0},
			{X: p.tox(t), Y: p.y},
		}, &diagram.Style{
			Stroke: color.Gray{0x80},
			Size:   1,
		})

		p.Text.Text(tick.String(), diagram.Point{
			X: p.tox(t),
			Y: 20,
		}, &diagram.Style{
			Stroke: color.Gray{0x40},
			Size:   12,
			Origin: diagram.Point{X: -1, Y: 0},
		})
	}
}

func (p *Plot) addPackage(t *Task) {
	if t.Duration() < p.IgnorePackage {
		return
	}

	r := diagram.R(
		p.tox(t.Start), p.y,
		p.tox(t.Finish), p.y+p.PackageHeight,
	)
	p.y += p.PackageHeight

	p.Spans.Rect(r, &diagram.Style{
		Fill: color.Gray{0xB0},
		Hint: t.Name + " " + t.Duration().String(),
	})

	p.Text.Text(t.Name, diagram.Point{
		X: r.Min.X + 5.0,
		Y: (r.Min.Y + r.Max.Y) / 2,
	}, &diagram.Style{
		Stroke: color.Black,
		Size:   r.Size().Y - 1,
		Origin: diagram.Point{X: -1, Y: 0},
	})

	p.Text.Text(t.Duration().Truncate(time.Second).String(), diagram.Point{
		X: r.Min.X - 5.0,
		Y: (r.Min.Y + r.Max.Y) / 2,
	}, &diagram.Style{
		Stroke: color.Black,
		Size:   r.Size().Y - 1,
		Origin: diagram.Point{X: 1, Y: 0},
	})

	p.line = 0
	attach := diagram.P(r.Min.X, r.Max.Y)
	for _, sub := range t.Sub {
		p.addTest(0, attach, sub)
	}
}

func (p *Plot) addTest(level int, parent diagram.Point, t *Task) {
	if t.Duration() < p.IgnoreTest {
		return
	}

	p.line++

	r := diagram.R(
		p.tox(t.Start), p.y,
		p.tox(t.Finish), p.y+p.TestHeight,
	)
	p.y += p.TestHeight

	hint := t.Name + " " + t.Duration().String()
	p.Spans.Rect(r, &diagram.Style{
		Fill: color.Gray{0xC0},
		Hint: hint,
	})
	p.drawEvents(level, r.Min.Y, r.Max.Y, t.Events, hint)

	p.Text.Text(t.Name, diagram.Point{
		X: r.Min.X + 2.0,
		Y: r.Max.Y - 2.0,
	}, &diagram.Style{
		Stroke: color.Black,
		Size:   r.Size().Y,
		Origin: diagram.Point{X: -1, Y: 1},
	})

	p.Assoc.Poly([]diagram.Point{
		{X: parent.X, Y: parent.Y},
		{X: parent.X, Y: (r.Min.Y + r.Max.Y) / 2},
		{X: r.Min.X, Y: (r.Min.Y + r.Max.Y) / 2},
	}, &diagram.Style{
		Stroke: color.Gray{0x30},
	})

	attach := diagram.P(r.Min.X, r.Max.Y)
	if len(t.Events) >= 3 {
		if t.Events[0].Action == ActionRun && t.Events[1].Action == ActionPause && t.Events[2].Action == ActionCont {
			attach.X = p.tox(t.Events[2].Time)
		}
	}

	for _, sub := range t.Sub {
		p.addTest(level+1, attach, sub)
	}
}

func (p *Plot) drawEvents(level int, top, bottom float64, events Events, hint string) {
	if len(events) <= 1 {
		return
	}

	style := &diagram.Style{Hint: hint}
	switch level % 3 {
	case 0:
		style.Fill = color.RGBA{R: byte(0x80 + 0x20*(p.line%2)), A: 0xFF}
	case 1:
		style.Fill = color.RGBA{G: byte(0x80 + 0x20*(p.line%2)), A: 0xFF}
	case 2:
		style.Fill = color.RGBA{B: byte(0x80 + 0x20*(p.line%2)), A: 0xFF}
	}

	prev := events[0]
	h := bottom - top
	for _, next := range events[1:] {
		ratio := 0.0
		if prev.Action == ActionPause {
			prev = next
			continue
		}
		r := diagram.R(
			p.tox(prev.Time), top+h*ratio,
			p.tox(next.Time), bottom-h*ratio,
		)
		p.Spans.Rect(r, style)

		prev = next
	}
}
