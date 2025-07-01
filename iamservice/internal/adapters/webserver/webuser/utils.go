package webuser

import (
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/samber/lo"
	"strconv"
)

func convertHeightToMm(feet, inches int) int {
	totalInches := float64(feet)*12 + float64(inches)
	return int(totalInches * 25.4)
}

func convertHeightMmToFeetInches(mm int) (int, int) {
	totalInches := float64(mm) / 25.4
	totalInchesRounded := int(totalInches + 0.5) // Round to nearest inch
	feet := totalInchesRounded / 12
	inches := totalInchesRounded % 12
	return feet, inches
}

func parseMetricHeightIntoMillimeters(str string) (*int, error) {
	if str == "" {
		return nil, nil
	}
	if height, err := strconv.ParseInt(str, 10, 32); err == nil {
		return lo.ToPtr(int(height) * 10), nil
	} else {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "Invalid height value")
	}
}

func parseImperialHeightIntoMillimeters(feetStr, inchesStr string) (*int, error) {
	if feetStr == "" || inchesStr == "" {
		return nil, nil
	}
	if feet, err := strconv.ParseInt(feetStr, 10, 32); err == nil {
		if inches, err := strconv.ParseInt(inchesStr, 10, 32); err == nil {
			heightMm := convertHeightToMm(int(feet), int(inches))
			return lo.ToPtr(heightMm), nil
		} else {
			return nil, katapp.NewErr(katapp.ErrInvalidInput, "Invalid inches value")
		}
	}
	return nil, katapp.NewErr(katapp.ErrInvalidInput, "Invalid feet value")
}

func parseMetricWeightIntoGrams(str string) (*int, error) {
	if str == "" {
		return nil, nil
	}
	if weightKg, err := strconv.ParseFloat(str, 32); err == nil {
		// Convert kg to grams (multiply by 1000) and round to nearest gram
		weightGrams := int(weightKg*1000 + 0.5)
		return &weightGrams, nil
	}
	return nil, katapp.NewErr(katapp.ErrInvalidInput, "Invalid weight value")
}

func parseImperialWeightIntoGrams(str string) (*int, error) {
	if str == "" {
		return nil, nil
	}
	if weightLbs, err := strconv.ParseFloat(str, 32); err == nil {
		// Convert lbs to kg, then to grams
		weightKg := weightLbs / 2.20462
		weightGrams := int(weightKg*1000 + 0.5) // Add 0.5 for proper rounding
		return &weightGrams, nil
	}
	return nil, katapp.NewErr(katapp.ErrInvalidInput, "Invalid weight value")
}
