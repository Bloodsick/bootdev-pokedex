package pokeapi

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Bloodisck/bootdev-pokedex/internal/pokecache"
)

// Client -
type Client struct {
	cache      pokecache.Cache
	httpClient http.Client
}

// NewClient -
func NewClient(timeout, cacheInterval time.Duration) Client {
	return Client{
		cache: pokecache.NewCache(cacheInterval),
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}

// LocationArea -
type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// LocationAreaResponse -
type LocationAreaResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []LocationArea `json:"results"`
}

// GetLocationAreas -
func (c *Client) GetLocationAreas(pageURL *string) (LocationAreaResponse, error) {
	url := "https://pokeapi.co/api/v2/location-area"
	if pageURL != nil {
		url = *pageURL
	}

	// 1. Check the cache first!
	if val, ok := c.cache.Get(url); ok {
		// Cache hit
		locationResp := LocationAreaResponse{}
		err := json.Unmarshal(val, &locationResp)
		if err != nil {
			return LocationAreaResponse{}, err
		}
		return locationResp, nil
	}

	// 2. Cache miss, make the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return LocationAreaResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return LocationAreaResponse{}, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return LocationAreaResponse{}, err
	}

	locationResp := LocationAreaResponse{}
	err = json.Unmarshal(dat, &locationResp)
	if err != nil {
		return LocationAreaResponse{}, err
	}

	// 3. Add to cache
	c.cache.Add(url, dat)

	return locationResp, nil
}
