package main

import (
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
)

type Sprites struct {
	Black     color.Color
	White     color.Color
	Red       color.Color
	Trans     color.Color
	PLASSPRITE int
	MINUSSPRITE int
	EXT       string
	img       *image.RGBA
	dir       string
	w, h      int
}

func NewSprites(spritesDir string, canvas image.Image) *Sprites {
	bounds := canvas.Bounds()
	img := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.Set(x, y, canvas.At(x, y))
		}
	}

	return &Sprites{
		Black:      color.RGBA{0, 0, 0, 255},
		White:      color.RGBA{255, 255, 255, 255},
		Red:        color.RGBA{255, 0, 0, 255},
		Trans:      color.RGBA{0, 0, 0, 0},
		PLASSPRITE: 10,
		MINUSSPRITE: 11,
		EXT:        ".png",
		img:       img,
		dir:       spritesDir,
		w:         bounds.Max.X,
		h:         bounds.Max.Y,
	}
}

func (s *Sprites) Dot(x, y int, color color.Color) {
	if y >= s.h || x >= s.w || y < 0 || x < 0 {
		return
	}
	s.img.Set(x, y, color)
}

func (s *Sprites) Draw(name string, index, xpos, ypos int) int {
	imageFilename := filepath.Join(s.dir, name+"_"+formatIndex(index)+s.EXT)
	img, err := loadImage(imageFilename)
	if err != nil {
		return 0
	}
	w, h := img.Bounds().Max.X, img.Bounds().Max.Y
	ypos -= h

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if xpos+x >= s.w || xpos+x < 0 {
				continue
			}
			if ypos+y >= s.h || ypos+y < 0 {
				continue
			}
			col := img.At(x, y)
			if col == s.Black {
				s.Dot(xpos+x, ypos+y, s.Black)
			} else if col == s.White {
				s.Dot(xpos+x, ypos+y, s.White)
			} else if col == s.Red {
				s.Dot(xpos+x, ypos+y, s.Black)
			}
		}
	}
	return w
}

func (s *Sprites) DrawInt(n, xpos, ypos int, isSign, isLeadZero bool) int {
	sign := s.PLASSPRITE
	if n < 0 {
		sign = s.MINUSSPRITE
	}
	n = abs(n)

	n1 := n / 10
	n2 := n % 10
	dx := 0

	if isSign {
		w := s.Draw("digit", sign, xpos+dx, ypos)
		dx += w + 1
	}
	if n1 != 0 || isLeadZero {
		w := s.Draw("digit", n1, xpos+dx, ypos)
		dx += w + 1
	}
	w := s.Draw("digit", n2, xpos+dx, ypos)
	dx += w + 1
	return dx
}

func (s *Sprites) DrawClock(xpos, ypos, h, m int) int {
	dx := 0
	w := s.DrawInt(h, xpos+dx, ypos, false, true)
	dx += w
	w = s.Draw("digit", 12, xpos+dx, ypos) // Assuming 12 is the index for ':'
	dx += w
	dx += s.DrawInt(m, xpos+dx, ypos, false, true)
	dx++
	return dx
}

func (s *Sprites) DrawCloud(percent, xpos, ypos, width, height int) {
	if percent < 2 {
		return
	}

	cloudSet := s.getCloudSet(percent)

	for _, c := range cloudSet {
		s.Draw("cloud", c, xpos+rand.Intn(width), ypos)
	}
}

func (s *Sprites) getCloudSet(percent int) []int {
	switch {
	case percent < 5:
		return []int{2}
	case percent < 10:
		return []int{3, 2}
	case percent < 20:
		return []int{5, 3, 2}
	case percent < 30:
		return []int{10, 5}
	case percent < 40:
		return []int{10, 10}
	case percent < 50:
		return []int{10, 10, 5}
	case percent < 60:
		return []int{30, 5}
	case percent < 70:
		return []int{30, 10}
	case percent < 80:
		return []int{30, 10, 5, 5}
	case percent < 90:
		return []int{30, 10, 10}
	default:
		return []int{50, 30, 10, 10, 5}
	}
}

func (s *Sprites) DrawRain(value, xpos, ypos, width int, tline []int) {
	ypos++
	r := 1.0 - (float64(value)/5.0)/20.0 // HEAVYRAIN and RAINFACTOR

	for x := xpos; x < xpos+width; x++ {
		for y := ypos; y < tline[x]; y += 2 {
			if x >= s.w || y >= s.h {
				continue
			}
			if rand.Float64() > r {
				s.img.Set(x, y, s.Black)
				s.img.Set(x, y-1, s.Black)
			}
		}
	}
}

func (s *Sprites) DrawSnow(value, xpos, ypos, width int, tline []int) {
	ypos++
	r := 1.0 - (float64(value)/5.0)/10.0 // HEAVYSNOW and SNOWFACTOR

	for x := xpos; x < xpos+width; x++ {
		for y := ypos; y < tline[x]; y += 2 {
			if x >= s.w || y >= s.h {
				continue
			}
			if rand.Float64() > r {
				s.img.Set(x, y, s.Black)
			}
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func formatIndex(index int) string {
	return fmt.Sprintf("%02d", index)
}

func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func main() {
	canvasFile := "../test.bmp"
	spritesDir := "../sprite"

	canvas, err := loadImage(canvasFile)
	if err != nil {
		panic(err)
	}

	s := NewSprites(spritesDir, canvas)
	s.Draw("house", 0, 100, 100)

	outputFile, err := os.Create("../tmp/sprites_test.bmp")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	png.Encode(outputFile, s.img)
}
