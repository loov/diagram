package diagram

import "image/color"

type Style struct {
	Stroke color.Color
	Fill   color.Color
	Size   Length

	// line only
	Dash       []Length
	DashOffset []Length

	// text only
	Font     string
	Rotation float64
	Origin   Point // {-1..1, -1..1}

	// SVG
	Hint  string
	Class string
}

func (style *Style) mustExist() {
	if style == nil {
		panic("style missing")
	}
}

func (style *Style) IsZero() bool {
	if style == nil {
		return true
	}

	return style.Stroke == nil && style.Fill == nil && style.Size == 0
}

func (style *Style) Or(other Style) *Style {
	if style.IsZero() {
		return &other
	}
	copy := *style
	return &copy
}
