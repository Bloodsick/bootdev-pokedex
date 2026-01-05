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
	Moves []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
	} `json:"moves"`
}

type Move struct {
	Name     string `json:"name"`
	Accuracy int    `json:"accuracy"`
	Power    int    `json:"power"`
	PP       int    `json:"pp"`
	Type     struct {
		Name string `json:"name"`
	} `json:"type"`
}

type PokemonSpecies struct {
	EvolutionChain struct {
		URL string `json:"url"`
	} `json:"evolution_chain"`
}

type EvolutionChainResponse struct {
	Chain ChainLink `json:"chain"`
}

type ChainLink struct {
	Species struct {
		Name string `json:"name"`
	} `json:"species"`
	EvolutionDetails []EvolutionDetail `json:"evolution_details"`
	EvolvesTo        []ChainLink       `json:"evolves_to"`
}

type EvolutionDetail struct {
	MinLevel *int `json:"min_level"`
	Item     *struct {
		Name string `json:"name"`
	} `json:"item"`
	Trigger struct {
		Name string `json:"name"`
	} `json:"trigger"`
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

func (c *Client) GetPokemonSpecies(name string) (PokemonSpecies, error) {
	url := "https://pokeapi.co/api/v2/pokemon-species/" + name

	if val, ok := c.cache.Get(url); ok {
		speciesResp := PokemonSpecies{}
		err := json.Unmarshal(val, &speciesResp)
		if err != nil {
			return PokemonSpecies{}, err
		}
		return speciesResp, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return PokemonSpecies{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PokemonSpecies{}, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return PokemonSpecies{}, err
	}

	speciesResp := PokemonSpecies{}
	err = json.Unmarshal(dat, &speciesResp)
	if err != nil {
		return PokemonSpecies{}, err
	}

	c.cache.Add(url, dat)
	return speciesResp, nil
}

func (c *Client) GetEvolutionChain(url string) (EvolutionChainResponse, error) {
	// 1. Check Cache
	if val, ok := c.cache.Get(url); ok {
		chainResp := EvolutionChainResponse{}
		err := json.Unmarshal(val, &chainResp)
		if err != nil {
			return EvolutionChainResponse{}, err
		}
		return chainResp, nil
	}

	// 2. Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return EvolutionChainResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return EvolutionChainResponse{}, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return EvolutionChainResponse{}, err
	}

	chainResp := EvolutionChainResponse{}
	err = json.Unmarshal(dat, &chainResp)
	if err != nil {
		return EvolutionChainResponse{}, err
	}

	// 3. Add to Cache
	c.cache.Add(url, dat)
	return chainResp, nil
}

func (c *Client) GetMove(name string) (Move, error) {
	url := "https://pokeapi.co/api/v2/move/" + name

	if val, ok := c.cache.Get(url); ok {
		moveResp := Move{}
		err := json.Unmarshal(val, &moveResp)
		if err != nil {
			return Move{}, err
		}
		return moveResp, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Move{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Move{}, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return Move{}, err
	}

	moveResp := Move{}
	err = json.Unmarshal(dat, &moveResp)
	if err != nil {
		return Move{}, err
	}

	c.cache.Add(url, dat)
	return moveResp, nil
}
