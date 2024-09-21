package main

import (
	"fmt"
	"math"
	"time"
)

type Sun struct {
	Lat      float64
	Long     float64
	TzOffset float64
	Day      int
	Time     float64
	SunriseT float64
	SunsetT  float64
	SolarNoonT float64
}

func NewSun(lat, long float64) *Sun {
	tzOffset := float64(time.Now().Sub(time.Now().UTC()).Seconds()) / 3600
	return &Sun{
		Lat:      lat,
		Long:     long,
		TzOffset: tzOffset,
	}
}

func (s *Sun) Sunrise(when time.Time) time.Time {
	s.prepTime(when)
	s.calc()
	return s.timeFromDecimalDay(s.SunriseT, when)
}

func (s *Sun) Sunset(when time.Time) time.Time {
	s.prepTime(when)
	s.calc()
	return s.timeFromDecimalDay(s.SunsetT, when)
}

func (s *Sun) SolarNoon(when time.Time) time.Time {
	s.prepTime(when)
	s.calc()
	return s.timeFromDecimalDay(s.SolarNoonT, when)
}

func (s *Sun) prepTime(when time.Time) {
	s.Day = int(when.Sub(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)).Hours()/24) + 40529
	h, m, sec := when.Clock()
	s.Time = float64(h)/24 + float64(m)/(24*60) + float64(sec)/(24*3600)
}

func (s *Sun) calc() {
	timezone := s.TzOffset
	latitude := s.Lat
	longitude := s.Long
	timeFrac := s.Time
	day := s.Day

	Jday := float64(day) + 2415018.5 + timeFrac - timezone/24
	Jcent := (Jday - 2451545) / 36525

	Manom := 357.52911 + Jcent*(35999.05029-0.0001537*Jcent)
	Mlong := math.Mod(280.46646+Jcent*(36000.76983+Jcent*0.0003032), 360)
	Eccent := 0.016708634 - Jcent*(0.000042037+0.0001537*Jcent)
	Mobliq := 23 + (26+(21.448-Jcent*(46.815+Jcent*(0.00059-Jcent*0.001813)))/60)/60
	obliq := Mobliq + 0.00256*math.Cos(degToRad(125.04-1934.136*Jcent))
	vary := math.Tan(degToRad(obliq / 2))
	vary = vary * vary

	Seqcent := math.Sin(degToRad(Manom))*(1.914602-Jcent*(0.004817+0.000014*Jcent)) +
		math.Sin(degToRad(2*Manom))*(0.019993-0.000101*Jcent) +
		math.Sin(degToRad(3*Manom))*0.000289

	Struelong := Mlong + Seqcent
	Sapplong := Struelong - 0.00569 - 0.00478*math.Sin(degToRad(125.04-1934.136*Jcent))
	declination := radToDeg(math.Asin(math.Sin(degToRad(obliq)) * math.Sin(degToRad(Sapplong))))

	eqtime := 4 * radToDeg(vary*math.Sin(2*degToRad(Mlong)) -
		2*Eccent*math.Sin(degToRad(Manom)) +
		4*Eccent*vary*math.Sin(degToRad(Manom))*math.Cos(2*degToRad(Mlong)) -
		0.5*vary*vary*math.Sin(4*degToRad(Mlong)) -
		1.25*Eccent*Eccent*math.Sin(2*degToRad(Manom)))

	hourangle := radToDeg(math.Acos(math.Cos(degToRad(90.833))/(math.Cos(degToRad(latitude))*math.Cos(degToRad(declination))) - math.Tan(degToRad(latitude))*math.Tan(degToRad(declination))))

	s.SolarNoonT = (720 - 4*longitude - eqtime + timezone*60) / 1440
	s.SunriseT = s.SolarNoonT - hourangle*4/1440
	s.SunsetT = s.SolarNoonT + hourangle*4/1440
}

func (s *Sun) timeFromDecimalDay(day float64, when time.Time) time.Time {
	hours := 24.0 * day
	h := int(hours)
	minutes := (hours - float64(h)) * 60
	m := int(minutes)
	seconds := (minutes - float64(m)) * 60
	sec := int(seconds)
	return time.Date(when.Year(), when.Month(), when.Day(), h, m, sec, 0, when.Location())
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}

func main() {
	s := NewSun(50.4546600, 30.5238000) // Default Kyiv
	now := time.Now()

	fmt.Println("Sunrise:", s.Sunrise(now))
	fmt.Println("Sunset:", s.Sunset(now))
	fmt.Println("Solar Noon:", s.SolarNoon(now))
}
