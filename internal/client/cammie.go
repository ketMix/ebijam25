package client

import "github.com/hajimehoshi/ebiten/v2"

type Cammie struct {
	x, y         float64
	zoom         float64 // We can probably "overload" this to implement LOD.
	locked       bool
	image        *ebiten.Image
	lastW, lastH int
	opts         ebiten.DrawImageOptions
}

func (c *Cammie) Setup() {
	c.zoom = 1.0
	c.image = ebiten.NewImage(1, 1) // Initialize with a dummy image.
	c.lastW, c.lastH = 1, 1
}

func (c *Cammie) Layout(ow, oh int) {
	if c.lastW == ow && c.lastH == oh {
		return
	}
	c.lastW = ow
	c.lastH = oh
	c.image = ebiten.NewImage(ow, oh)
}

func (c *Cammie) Update() {
	c.opts.GeoM.Reset()
	c.opts.GeoM.Translate(-c.x, -c.y)
	c.opts.GeoM.Translate(float64(c.image.Bounds().Dx())/2, float64(c.image.Bounds().Dy())/2)
	c.opts.GeoM.Scale(c.zoom, c.zoom)
}

func (c *Cammie) Draw(screen *ebiten.Image) {
	// NOTE: We could implement screenshake effects here.
	screen.DrawImage(c.image, nil)
}

func (c *Cammie) SetPosition(x, y float64) {
	// We could interpolate from this value instead of setting directly..
	c.x = x
	c.y = y
}

func (c *Cammie) AddPosition(dx, dy float64) {
	c.x += dx
	c.y += dy
}

func (c *Cammie) SetZoom(zoom float64) {
	c.zoom = zoom
}

func (c *Cammie) Locked() bool {
	return c.locked
}

func (c *Cammie) ToggleLocked() {
	c.locked = !c.locked
}
