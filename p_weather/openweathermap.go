package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

const (
    KTOC               = 273.15
    FORECAST_PERIOD_HOURS = 3
    OWMURL             = "http://api.openweathermap.org/data/2.5/"
    FILENAME_CURR      = "openweathermap_curr_"
    FILENAME_FORECAST  = "openweathermap_fcst_"
    FILENAME_EXT       = ".json"
    FILETOOOLD_SEC     = 15 * 60 // 15 mins
    TOOMUCHTIME_SEC    = 4 * 60 * 60 // 4 hours 
)

type WeatherInfo struct {
    T         time.Time
    ID        int
    Clouds    int
    Rain      float64
    Snow      float64
    Windspeed float64
    Winddeg   float64
    Temp      float64
}

func NewWeatherInfo(fdata map[string]interface{}) (*WeatherInfo, error) {
    t := time.Unix(int64(fdata["dt"].(float64)), 0)
    id := int(fdata["weather"].([]interface{})[0].(map[string]interface{})["id"].(float64))

    clouds := 0
    if val, ok := fdata["clouds"].(map[string]interface{})["all"]; ok {
        clouds = int(val.(float64))
    }

    rain := 0.0
    if val, ok := fdata["rain"].(map[string]interface{})["3h"]; ok {
        rain = val.(float64)
    }

    snow := 0.0
    if val, ok := fdata["snow"].(map[string]interface{})["3h"]; ok {
        snow = val.(float64)
    }

    windspeed := 0.0
    if val, ok := fdata["wind"].(map[string]interface{})["speed"]; ok {
        windspeed = val.(float64)
    }

    winddeg := 0.0
    if val, ok := fdata["wind"].(map[string]interface{})["deg"]; ok {
        winddeg = val.(float64)
    }

    temp := fdata["main"].(map[string]interface{})["temp"].(float64) - KTOC

    return &WeatherInfo{T: t, ID: id, Clouds: clouds, Rain: rain, Snow: snow, Windspeed: windspeed, Winddeg: winddeg, Temp: temp}, nil
}

func (w *WeatherInfo) Print() {
    fmt.Printf("%s %d %03d%% %.2f %.2f %+.2f (%5.1f,%03d)\n",
        w.T, w.ID, w.Clouds, w.Rain, w.Snow, w.Temp, w.Windspeed, int(w.Winddeg))
}

type OpenWeatherMap struct {
    Latitude  float64
    Longitude float64
    Rootdir   string
    F         []*WeatherInfo
    URL_FORECAST string
    URL_CURR      string
    PLACEKEY      string
}

func NewOpenWeatherMap(apikey string, latitude, longitude float64, rootdir string) *OpenWeatherMap {
    owm := &OpenWeatherMap{
        Latitude:  latitude,
        Longitude: longitude,
        Rootdir:   rootdir,
    }
    owm.URL_FORECAST = fmt.Sprintf("%sforecast?lat=%.4f&lon=%.4f&mode=json&APPID=%s", OWMURL, latitude, longitude, apikey)
    owm.URL_CURR = fmt.Sprintf("%sweather?lat=%.4f&lon=%.4f&mode=json&APPID=%s", OWMURL, latitude, longitude, apikey)

    if _, err := os.Stat(owm.Rootdir); os.IsNotExist(err) {
        os.Mkdir(owm.Rootdir, os.ModePerm)
    }

    owm.PLACEKEY = owm.makePlaceKey()

    return owm
}

func (owm *OpenWeatherMap) makePlaceKey() string {
    return makeCoordinateKey(owm.Latitude) + makeCoordinateKey(owm.Longitude)
}

func makeCoordinateKey(p float64) string {
    n := int(p * 10000)
    return fmt.Sprintf("%08X", n&0xFFFFFFFF)[2:]
}

func (owm *OpenWeatherMap) fetchFromWWW() error {
    resp, err := http.Get(owm.URL_FORECAST)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    fjsontext, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    
    ioutil.WriteFile(filepath.Join(owm.Rootdir, FILENAME_FORECAST + owm.PLACEKEY + FILENAME_EXT), fjsontext, 0644)

    var fdata map[string]interface{}
    if err := json.Unmarshal(fjsontext, &fdata); err != nil {
        return err
    }

    resp, err = http.Get(owm.URL_CURR)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    cjsontext, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }

    ioutil.WriteFile(filepath.Join(owm.Rootdir, FILENAME_CURR + owm.PLACEKEY + FILENAME_EXT), cjsontext, 0644)

    var cdata map[string]interface{}
    if err := json.Unmarshal(cjsontext, &cdata); err != nil {
        return err
    }

    return owm.fromJSON(cdata, fdata)
}

func (owm *OpenWeatherMap) fromJSON(data_curr, data_fcst map[string]interface{}) error {
    owm.F = nil
    f, err := NewWeatherInfo(data_curr)
    if err != nil {
        return err
    }
    owm.F = append(owm.F, f)

    list, ok := data_fcst["list"].([]interface{})
    if !ok {
        return fmt.Errorf("no forecast data available")
    }

    for _, fdata := range list {
        if info, err := NewWeatherInfo(fdata.(map[string]interface{})); err == nil {
            owm.F = append(owm.F, info)
        }
    }
    return nil
}

func (owm *OpenWeatherMap) isFileTooOld(filename string) bool {
    fileInfo, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return true
    }
    return time.Since(fileInfo.ModTime()).Seconds() > FILETOOOLD_SEC
}

func (owm *OpenWeatherMap) fromAuto() error {
    filenameForecast := filepath.Join(owm.Rootdir, FILENAME_FORECAST + owm.PLACEKEY + FILENAME_EXT)
    filenameCurr := filepath.Join(owm.Rootdir, FILENAME_CURR + owm.PLACEKEY + FILENAME_EXT)

    if owm.isFileTooOld(filenameForecast) || owm.isFileTooOld(filenameCurr) {
        fmt.Println("Using WWW")
        return owm.fetchFromWWW()
    }

    fmt.Printf("Using Cache '%s', '%s'\n", filenameCurr, filenameForecast)
    return owm.fromFile()
}

func (owm *OpenWeatherMap) fromFile() error {
    forecastFile := filepath.Join(owm.Rootdir, FILENAME_FORECAST + owm.PLACEKEY + FILENAME_EXT)
    currentFile := filepath.Join(owm.Rootdir, FILENAME_CURR + owm.PLACEKEY + FILENAME_EXT)

    fdata, err := os.ReadFile(forecastFile)
    if err != nil {
        return err
    }

    cdata, err := os.ReadFile(currentFile)
    if err != nil {
        return err
    }

    var forecast map[string]interface{}
    if err := json.Unmarshal(fdata, &forecast); err != nil {
        return err
    }

    var current map[string]interface{}
    if err := json.Unmarshal(cdata, &current); err != nil {
        return err
    }

    return owm.fromJSON(current, forecast)
}

func (owm *OpenWeatherMap) getCurr() *WeatherInfo {
    if len(owm.F) == 0 {
        return nil
    }
    return owm.F[0]
}

func (owm *OpenWeatherMap) get(time time.Time) *WeatherInfo {
    for _, f := range owm.F {
        if f.T.After(time) {
            return f
        }
    }
    return nil
}

func (owm *OpenWeatherMap) printAll() {
    for _, f := range owm.F {
        f.Print()
    }
}

func main() {
    // Example usage
    apikey := "your_api_key"
    latitude := -33.8651
    longitude := 151.2099
    rootdir := "./weather_data"

    owm := NewOpenWeatherMap(apikey, latitude, longitude, rootdir)

    if err := owm.fromAuto(); err != nil {
        fmt.Println("Error:", err)
    } else {
        owm.printAll()
    }
}
