package game

import "github.com/Bloodisck/bootdev-pokedex/internal/pokeapi"

func FindNextEvolution(chain pokeapi.ChainLink, currentName string) (nextName string, level int, item string) {
	if chain.Species.Name == currentName {
		if len(chain.EvolvesTo) > 0 {
			nextLink := chain.EvolvesTo[0]

			nextName = nextLink.Species.Name

			if len(nextLink.EvolutionDetails) > 0 {
				details := nextLink.EvolutionDetails[0]
				if details.MinLevel != nil {
					level = *details.MinLevel
				}
				if details.Item != nil {
					item = details.Item.Name
				}
			}
			return
		}
		return "", 0, ""
	}

	for _, child := range chain.EvolvesTo {
		n, l, i := FindNextEvolution(child, currentName)
		if n != "" {
			return n, l, i
		}
	}

	return "", 0, ""
}
