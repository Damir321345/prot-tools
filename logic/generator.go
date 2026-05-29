package logic

import (
	"math/rand"
	"strconv"
	"strings"
)

func GenerateCurrents(ratio string, class string) (primary float64, secondary float64, err error) {
	parts := strings.Split(ratio, "/")

	prim, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	sec, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	ratioValue := float64(prim) / float64(sec)
	deviation := (rand.Float64()*2 - 1) * 0.01
	secondary = ratioValue * (1 + deviation)
	primary = float64(prim)
	return primary, secondary, nil
}
