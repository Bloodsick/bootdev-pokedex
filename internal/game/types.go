package game

var typeChart = map[string]map[string]float64{
	"fire": {
		"grass": 2.0, "ice": 2.0, "bug": 2.0,
		"water": 0.5, "rock": 0.5, "fire": 0.5,
	},
	"water": {
		"fire": 2.0, "ground": 2.0, "rock": 2.0,
		"water": 0.5, "grass": 0.5,
	},
	"grass": {
		"water": 2.0, "ground": 2.0, "rock": 2.0,
		"fire": 0.5, "grass": 0.5, "flying": 0.5,
	},
	"electric": {
		"water": 2.0, "flying": 2.0,
		"ground": 0.5, "grass": 0.5, "electric": 0.5,
	},
	"normal": {
		"ghost": 0.5, "rock": 0.5,
	},
}

func GetTypeEffectiveness(moveType string, defenderTypes []string) float64 {
	multiplier := 1.0

	if chart, ok := typeChart[moveType]; ok {
		for _, defType := range defenderTypes {
			if mod, found := chart[defType]; found {
				multiplier *= mod
			}
		}
	}
	return multiplier
}
