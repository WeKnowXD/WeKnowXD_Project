package weather

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	cache      map[string]any
	cachedAt   time.Time
	cacheMutex sync.Mutex
)

const cacheDuration = 30 * time.Minute

func Get() (map[string]any, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cache == nil || time.Since(cachedAt) > cacheDuration {
		log.Println("func: apiWeather, cache empty or expired, fetching new forecast")

		data, err := fetch()
		if err != nil {
			log.Println("func: apiWeather, fetching weather failed:", err)

			if cache == nil {
				return nil, err
			}
		} else {
			cache = data
			cachedAt = time.Now()
		}
	}

	return cache, nil
}

func fetch() (map[string]any, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := "https://api.weatherapi.com/v1/forecast.json?key=" + apiKey + "&q=Copenhagen&days=3"

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weatherapi returned status %d", resp.StatusCode)
	}

	var weatherData map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, err
	}

	return weatherData, nil
}
