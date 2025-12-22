package charts

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type Sparkline struct {
	widget.BaseWidget

	mu        sync.RWMutex
	values    []float64
	maxPoints int

	lineColor color.Color
	bgColor   color.Color
}

func NewSparkline(maxPoints int, lineColor, bgColor color.Color) *Sparkline {
	if maxPoints <= 0 {
		maxPoints = 60
	}
	s := &Sparkline{
		maxPoints: maxPoints,
		lineColor: lineColor,
		bgColor:   bgColor,
	}
	s.ExtendBaseWidget(s)
	return s
}

func (s *Sparkline) Push(value float64) {
	s.mu.Lock()
	s.values = append(s.values, value)
	if len(s.values) > s.maxPoints {
		s.values = s.values[len(s.values)-s.maxPoints:]
	}
	s.mu.Unlock()
	s.Refresh()
}

func (s *Sparkline) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	img.FillMode = canvas.ImageFillStretch

	r := &sparklineRenderer{
		sparkline: s,
		img:       img,
	}
	r.redraw(fyne.NewSize(1, 1))
	return r
}

type sparklineRenderer struct {
	sparkline *Sparkline
	img       *canvas.Image
	rgba      *image.RGBA
	size      fyne.Size
}

func (r *sparklineRenderer) Destroy() {}

func (r *sparklineRenderer) Layout(size fyne.Size) {
	if size.Width <= 0 || size.Height <= 0 {
		size = fyne.NewSize(1, 1)
	}
	r.redraw(size)
	r.img.Resize(size)
}

func (r *sparklineRenderer) MinSize() fyne.Size {
	return fyne.NewSize(240, 72)
}

func (r *sparklineRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img}
}

func (r *sparklineRenderer) Refresh() {
	r.redraw(r.size)
	r.img.Refresh()
}

func (r *sparklineRenderer) redraw(size fyne.Size) {
	r.size = size
	w := int(math.Max(1, float64(size.Width)))
	h := int(math.Max(1, float64(size.Height)))

	if r.rgba == nil || r.rgba.Bounds().Dx() != w || r.rgba.Bounds().Dy() != h {
		r.rgba = image.NewRGBA(image.Rect(0, 0, w, h))
		r.img.Image = r.rgba
	}

	draw.Draw(r.rgba, r.rgba.Bounds(), &image.Uniform{C: r.sparkline.bgColor}, image.Point{}, draw.Src)

	r.sparkline.mu.RLock()
	values := append([]float64(nil), r.sparkline.values...)
	lineColor := r.sparkline.lineColor
	r.sparkline.mu.RUnlock()

	if len(values) < 2 {
		return
	}

	minVal, maxVal := values[0], values[0]
	for _, v := range values[1:] {
		minVal = math.Min(minVal, v)
		maxVal = math.Max(maxVal, v)
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	leftPad := 2.0
	rightPad := 2.0
	topPad := 2.0
	bottomPad := 2.0

	plotW := math.Max(1, float64(w)-leftPad-rightPad)
	plotH := math.Max(1, float64(h)-topPad-bottomPad)

	toPoint := func(i int, v float64) (int, int) {
		x := leftPad + (float64(i)/float64(len(values)-1))*plotW
		norm := (v - minVal) / (maxVal - minVal)
		y := topPad + (1-norm)*plotH
		return int(x), int(y)
	}

	// Simple polyline draw (Bresenham).
	for i := 0; i < len(values)-1; i++ {
		x0, y0 := toPoint(i, values[i])
		x1, y1 := toPoint(i+1, values[i+1])
		drawLine(r.rgba, x0, y0, x1, y1, lineColor)
	}
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx := int(math.Abs(float64(x1 - x0)))
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -int(math.Abs(float64(y1 - y0)))
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy

	for {
		if x0 >= 0 && y0 >= 0 && x0 < img.Bounds().Dx() && y0 < img.Bounds().Dy() {
			img.Set(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}
