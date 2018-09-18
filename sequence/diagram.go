package sequence

import (
	"image/color"
	"math"
	"sort"
	"strings"

	"github.com/loov/diagram"
)

type Diagram struct {
	Start, End Time

	Lanes    []*Lane
	Messages []*Message

	AutoSleep Time
	AutoDelay Time

	Theme struct {
		TimeScale     diagram.Length // length per time-unit
		CaptionHeight diagram.Length
		LaneWidth     diagram.Length
		LanePadding   diagram.Length

		Time    diagram.Style
		Caption diagram.Style
		Message diagram.Style
		Send    diagram.Style
	}
}

func New() *Diagram {
	dia := &Diagram{}

	dia.Start, dia.End = Automatic, Automatic

	dia.AutoSleep = 0.5
	dia.AutoDelay = 0.5

	const fontSize = 12
	const lineHeight = 16

	dia.Theme.TimeScale = lineHeight * 2
	dia.Theme.CaptionHeight = lineHeight * 3
	dia.Theme.LaneWidth = 200
	dia.Theme.LanePadding = lineHeight

	dia.Theme.Caption = diagram.Style{
		Stroke: nil,
		Fill:   color.NRGBA{0, 0, 0, 255},
		Size:   16,
	}
	dia.Theme.Time = diagram.Style{
		Stroke: color.NRGBA{30, 30, 30, 255},
		Size:   1,
		Dash:   []diagram.Length{4},
	}
	dia.Theme.Message = diagram.Style{
		Stroke: nil,
		Fill:   color.NRGBA{0, 0, 0, 255},
		Size:   12,
	}
	dia.Theme.Send = diagram.Style{
		Stroke: color.NRGBA{0, 0, 0, 255},
		Size:   1.3,
		// TODO: arrow
	}

	return dia
}

type Role = string

type Lane struct {
	Order int
	Name  Role

	Start Time
	End   Time

	Center diagram.Length

	Caption diagram.Style
	Line    diagram.Style
}

type Message struct {
	From Role
	To   Role
	Text string

	When  Time
	Sleep Time
	Delay Time

	Caption diagram.Style
	Line    diagram.Style
}

func FromTo(from, to Role, message string) *Message {
	return &Message{
		From:  from,
		To:    to,
		Text:  message,
		When:  Automatic,
		Sleep: Automatic,
		Delay: Automatic,
	}
}

func (message *Message) Start() Time { return message.When }
func (message *Message) End() Time   { return message.When + message.Delay }

func (message *Message) At(at Time) *Message           { message.When = at; return message }
func (message *Message) WithSleep(sleep Time) *Message { message.Sleep = sleep; return message }
func (message *Message) WithDelay(delay Time) *Message { message.Delay = delay; return message }

func (dia *Diagram) Add(messages ...*Message) {
	dia.Messages = append(dia.Messages, messages...)
}

func (dia *Diagram) Lane(name string) *Lane {
	for _, lane := range dia.Lanes {
		if strings.EqualFold(lane.Name, name) {
			return lane
		}
	}

	lane := &Lane{}
	lane.Name = name
	lane.Order = len(dia.Lanes)
	lane.Start = Automatic
	lane.End = Automatic

	dia.Lanes = append(dia.Lanes, lane)

	return lane
}

func (dia *Diagram) normalize() {
	dia.normalizeTimes()
	dia.normalizeLanes()
}

func (dia *Diagram) normalizeTimes() {
	last := &Message{}
	for _, message := range dia.Messages {
		if IsAutomatic(message.When) {
			if !IsAutomatic(message.Sleep) {
				message.When = last.End() + message.Sleep
			} else {
				message.When = last.End() + dia.AutoSleep
			}
		}
		if IsAutomatic(message.Delay) {
			message.Delay = dia.AutoDelay
		}
		last = message
	}

	sort.Slice(dia.Messages, func(i, k int) bool {
		a, b := dia.Messages[i], dia.Messages[i]
		if a.When == b.When {
			return a.Delay < b.Delay
		}
		return a.When < b.When
	})
}

func (dia *Diagram) normalizeLanes() {
	for _, message := range dia.Messages {
		dia.Start = Min(dia.Start, message.Start())
		dia.End = Max(dia.End, message.End())

		from := dia.Lane(message.From)
		from.Start = Min(from.Start, message.Start())
		from.End = Max(from.End, message.End())

		to := dia.Lane(message.To)
		to.Start = Min(to.Start, message.Start())
		to.End = Max(to.End, message.End())
	}
}

func (dia *Diagram) Size() (width, height float64) {
	dia.normalize()

	width = float64(len(dia.Lanes)) * dia.Theme.LaneWidth
	height = dia.Theme.CaptionHeight + 2*dia.Theme.LanePadding + (dia.End-dia.Start)*dia.Theme.TimeScale
	return width, height
}

func (dia *Diagram) Draw(canvas diagram.Canvas) {
	dia.normalize()

	guide := canvas.Layer(-1)
	sends := canvas.Layer(0)
	texts := canvas.Layer(1)

	y0 := dia.Theme.CaptionHeight + dia.Theme.LanePadding
	y1 := y0 + (dia.End-dia.Start)*dia.Theme.TimeScale + dia.Theme.LanePadding
	for _, lane := range dia.Lanes {
		lane.Center = (float64(lane.Order) + 0.5) * dia.Theme.LaneWidth

		guide.Poly(diagram.Ps(lane.Center, y0-dia.Theme.LanePadding, lane.Center, y1),
			lane.Line.Or(dia.Theme.Time))

		texts.Text(lane.Name, diagram.P(lane.Center, dia.Theme.CaptionHeight*0.5),
			lane.Caption.Or(dia.Theme.Caption))
	}

	for _, message := range dia.Messages {
		fromx := dia.Lane(message.From).Center
		fromy := y0 + (message.Start()-dia.Start)*dia.Theme.TimeScale
		tox := dia.Lane(message.To).Center
		toy := y0 + (message.End()-dia.Start)*dia.Theme.TimeScale

		lineStyle := message.Line.Or(dia.Theme.Send)
		sends.Poly(diagram.Ps(fromx, fromy, tox, toy), lineStyle)

		dx, dy := tox-fromx, toy-fromy
		angle := math.Atan2(dy, dx)

		var s = lineStyle.Size * 4
		var sn, cs float64
		sn, cs = math.Sincos(angle - math.Pi + math.Pi/8)
		sends.Poly(diagram.Ps(tox, toy, tox+cs*s, toy+sn*s), lineStyle)
		sn, cs = math.Sincos(angle - math.Pi - math.Pi/8)
		sends.Poly(diagram.Ps(tox, toy, tox+cs*s, toy+sn*s), lineStyle)

		if message.Text != "" {
			textstyle := message.Caption.Or(dia.Theme.Message)

			if dx < 0 {
				// right-to-left
				angle = math.Atan2(-dy, -dx)
			}

			textstyle.Rotation = angle
			texts.Text(message.Text,
				diagram.P((fromx+tox)*0.5, (fromy+toy)*0.5-textstyle.Size*0.6), textstyle)
		}
	}
}
