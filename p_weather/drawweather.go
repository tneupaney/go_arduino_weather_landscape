package main

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"time"
)

type DrawWeather struct {
	XSTART              int
	XSTEP               int
	XFLAT               int
	YSTEP               int
	DEFAULT_DEGREE_PER_PIXEL float64

	img      image.Image
	sprite   *Sprites
	IMGEWIDTH, IMGHEIGHT int
	tmin, tmax, temprange float64
	degreeperpixel        float64
	ypos                  int
}

func NewDrawWeather(canvas image.Image, sprites *Sprites) *DrawWeather {
	return &DrawWeather{
		XSTART:              32,
		XSTEP:               44,
		XFLAT:               10,
		YSTEP:               50,
		DEFAULT_DEGREE_PER_PIXEL: 0.5,
		img:      canvas,
		sprite:   sprites,
		IMGEWIDTH:  canvas.Bounds().Dx(),
		IMGHEIGHT: canvas.Bounds().Dy(),
	}
}

func (dw *DrawWeather) mybezier(x, xa, ya, xb, yb float64) int {
	xc := (xb + xa) / 2.0
	d := xb - xa
	t := (x - xa) / d
	return int(dw.mybeizelfnc(t, ya, ya, yb, yb))
}

func (dw *DrawWeather) mybeizelfnc(t, d0, d1, d2, d3 float64) float64 {
	return (1-t)*((1-t)*((1-t)*d0+t*d1) + t*((1-t)*d1+t*d2)) + t*((1-t)*((1-t)*d1+t*d2)+t*((1-t)*d2+t*d3)))
}

func (dw *DrawWeather) TimeDiffToPixels(dt time.Duration) int {
	ds := dt.Seconds()
	secondsPerPixel := float64(WeatherInfo.FORECAST_PERIOD_HOURS*60*60) / float64(dw.XSTEP)
	return int(ds / secondsPerPixel)
}

func (dw *DrawWeather) DegToPix(t float64) int {
	n := (t - dw.tmin) / dw.degreeperpixel
	y := dw.ypos + dw.YSTEP - int(n)
	return y
}

func (dw *DrawWeather) Draw(ypos int, owm *OpenWeatherMap) {
	dw.ypos = ypos
	nForecast := (dw.IMGEWIDTH - dw.XSTART) / dw.XSTEP
	maxTime := time.Now().Add(time.Duration(WeatherInfo.FORECAST_PERIOD_HOURS*nForecast) * time.Hour)

	dw.tmin, dw.tmax = owm.GetTempRange(maxTime)
	dw.temprange = dw.tmax - dw.tmin

	if dw.temprange < float64(dw.YSTEP) {
		dw.degreeperpixel = dw.DEFAULT_DEGREE_PER_PIXEL
	} else {
		dw.degreeperpixel = dw.temprange / float64(dw.YSTEP)
	}

	tline := make([]int, dw.IMGEWIDTH+dw.XSTEP+1)
	f := owm.GetCurr()
	oldTemp := f.Temp
	oldY := dw.DegToPix(oldTemp)
	for i := 0; i < dw.XSTART; i++ {
		tline[i] = oldY
	}
	yClouds := int(ypos - dw.YSTEP/2)
	f.Print()

	dw.sprite.Draw("house", 0, 0, oldY)
	dw.sprite.DrawInt(oldTemp, 8, oldY+10)
	dw.sprite.DrawCloud(f.Clouds, 0, yClouds, dw.XSTART, dw.YSTEP/2)
	dw.sprite.DrawRain(f.Rain, 0, yClouds, dw.XSTART, tline)
	dw.sprite.DrawSnow(f.Snow, 0, yClouds, dw.XSTART, tline)

	t := time.Now()
	dt := time.Duration(WeatherInfo.FORECAST_PERIOD_HOURS) * time.Hour
	tf := t

	xpos := dw.XSTART
	nForecast = int(nForecast)

	n := (dw.XSTEP - dw.XFLAT) / 2
	for i := 0; i <= nForecast; i++ {
		f = owm.Get(tf)
		if f == nil {
			continue
		}
		f.Print()
		newTemp := f.Temp
		newY := dw.DegToPix(newTemp)
		for j := 0; j < n; j++ {
			tline[xpos+j] = dw.mybezier(float64(xpos+j), float64(xpos), float64(oldY), float64(xpos+n), float64(newY))
		}

		for j := 0; j < dw.XFLAT; j++ {
			tline[xpos+j+n] = newY
		}

		xpos += n + dw.XFLAT
		n = (dw.XSTEP - dw.XFLAT)
		oldTemp = newTemp
		oldY = newY
		tf = tf.Add(dt)
	}

	s := sun{Lat: owm.Lat, Lon: owm.Lon}
	tf = t
	xpos = dw.XSTART
	objCounter := 0
	for i := 0; i <= nForecast; i++ {
		f = owm.Get(tf)
		if f == nil {
			continue
		}

		tSunrise := s.Sunrise(tf)
		tSunset := s.Sunset(tf)

		yMoon := ypos - dw.YSTEP*5/8

		if tf.Before(tSunrise) && tf.Add(dt).After(tSunrise) {
			dx := dw.TimeDiffToPixels(tSunrise.Sub(tf)) - dw.XSTEP/2
			dw.sprite.Draw("sun", 0, xpos+dx, yMoon)
			objCounter++
			if objCounter == 2 {
				break
			}
		}

		if tf.Before(tSunset) && tf.Add(dt).After(tSunset) {
			dx := dw.TimeDiffToPixels(tSunset.Sub(tf)) - dw.XSTEP/2
			dw.sprite.Draw("moon", 0, xpos+dx, yMoon)
			objCounter++
			if objCounter == 2 {
				break
			}
		}

		xpos += dw.XSTEP
		tf = tf.Add(dt)
	}

	isTminPrinted := false
	isTmaxPrinted := false
	tf = t
	xpos = dw.XSTART
	n = (dw.XSTEP - dw.XFLAT) / 2
	for i := 0; i <= nForecast; i++ {
		f = owm.Get(tf)
		if f == nil {
			continue
		}

		yClouds := int(ypos - dw.YSTEP/2)

		if f.Temp == dw.tmin && !isTminPrinted {
			dw.sprite.DrawInt(f.Temp, xpos+n, tline[xpos+n]+10)
			isTminPrinted = true
		}

		if f.Temp == dw.tmax && !isTmaxPrinted {
			dw.sprite.DrawInt(f.Temp, xpos+n, tline[xpos+n]+10)
			isTmaxPrinted = true
		}

		t0 := f.T.Add(-dt / 2)
		t1 := f.T.Add(dt / 2)

		dtOneHour := time.Duration(time.Hour)
		dxOneHour := float64(dw.XSTEP) / float64(WeatherInfo.FORECAST_PERIOD_HOURS)
		tt := t0
		xx := float64(xpos)
		for tt.Before(t1) {
			ix := int(xx)
			if tt.Hour() == 12 {
				dw.sprite.Draw("flower", 1, ix, tline[ix])
			}
			if tt.Hour() == 0 {
				dw.sprite.Draw("flower", 0, ix, tline[ix])
			}
			if tt.Hour() == 6 || tt.Hour() == 18 || tt.Hour() == 3 || tt.Hour() == 15 || tt.Hour() == 9 || tt.Hour() == 21 {
				dw.sprite.DrawWind(f.WindSpeed, f.WindDeg, ix, tline)
			}

			tt = tt.Add(dtOneHour)
			xx += dxOneHour
		}

		dw.sprite.DrawCloud(f.Clouds, xpos, yClouds, dw.XSTEP, dw.YSTEP/2)
		dw.sprite.DrawRain(f.Rain, xpos, yClouds, dw.XSTEP, tline)
		dw.sprite.DrawSnow(f.Snow, xpos, yClouds, dw.XSTEP, tline)

		xpos += dw.XSTEP
		tf = tf.Add(dt)
	}

	black := 0
	for x := 0; x < dw.IMGEWIDTH; x++ {
		if tline[x] < dw.IMGHEIGHT {
			dw.sprite.Dot(x, tline[x], black)
		} else {
			fmt.Printf("out of range: %d - %d(max %d)\n", x, tline[x], dw.IMGHEIGHT)
		}
	}
}

func main() {
	// Assuming you have methods to create your canvas and sprites
	canvas := createCanvas() // Implement your canvas creation logic
	sprites := loadSprites()  // Implement your sprite loading logic

	dw := NewDrawWeather(canvas, sprites)
	owm := NewOpenWeatherMap() // Implement your OpenWeatherMap initialization logic

	dw.Draw(100, owm) // Replace 100 with your desired ypos
}
