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

type LocationAreaDetail struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
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

func (c *Client) GetLocationArea(locationAreaName string) (LocationAreaDetail, error) {
	url := "https://pokeapi.co/api/v2/location-area/" + locationAreaName

	// 1. Check the cache
	if val, ok := c.cache.Get(url); ok {
		locationDetail := LocationAreaDetail{}
		err := json.Unmarshal(val, &locationDetail)
		if err != nil {
			return LocationAreaDetail{}, err
		}
		return locationDetail, nil
	}

	// 2. Cache miss
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return LocationAreaDetail{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return LocationAreaDetail{}, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return LocationAreaDetail{}, err
	}

	locationDetail := LocationAreaDetail{}
	err = json.Unmarshal(dat, &locationDetail)
	if err != nil {
		return LocationAreaDetail{}, err
	}

	// 3. Add to cache
	c.cache.Add(url, dat)

	return locationDetail, nil
}

func (c *Client) GetPokemon(pokemonName string) (Pokemon, error) {
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemonName

	if val, ok := c.cache.Get(url); ok {
		pokemonResp := Pokemon{}
		err := json.Unmarshal(val, &pokemonResp)
		if err != nil {
			return Pokemon{}, err
		}
		return pokemonResp, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Pokemon{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Pokemon{}, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return Pokemon{}, err
	}

	pokemonResp := Pokemon{}
	err = json.Unmarshal(dat, &pokemonResp)
	if err != nil {
		return Pokemon{}, err
	}

	c.cache.Add(url, dat)

	return pokemonResp, nil
}
