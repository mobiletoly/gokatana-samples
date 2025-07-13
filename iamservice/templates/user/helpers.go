package user

import (
	"fmt"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/oapi-codegen/runtime/types"
	"time"
)

// ============================================================================
// FORMATTING HELPER FUNCTIONS
// ============================================================================

func formatDate(date types.Date) string {
	return date.Format("January 2, 2006")
}

func formatDateTime(dateTime time.Time) string {
	return dateTime.Format("January 2, 2006 at 3:04 PM")
}

func formatInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}

func formatDateForInput(date types.Date) string {
	return date.Format("2006-01-02")
}

func formatGender(gender swagger.UserProfileGender) string {
	switch gender {
	case swagger.Male:
		return "Male"
	case swagger.Female:
		return "Female"
	case swagger.Other:
		return "Other"
	default:
		return "Other"
	}
}

func formatWeightFromGrams(weightGrams int) string {
	weightKg := float64(weightGrams) / 1000.0
	return fmt.Sprintf("%.2f", weightKg)
}

func formatHeightByPreference(heightMm int, isMetric bool) string {
	if isMetric {
		cm := float64(heightMm) / 10.0
		return fmt.Sprintf("%.0f cm", cm)
	} else {
		totalInches := float64(heightMm) / 25.4
		totalInchesRounded := int(totalInches + 0.5) // Round to nearest inch
		feetPart := totalInchesRounded / 12
		inchesPart := totalInchesRounded % 12
		return fmt.Sprintf("%d'%d\"", feetPart, inchesPart)
	}
}

func formatWeightByPreference(weightGrams int, isMetric bool) string {
	if isMetric {
		return formatWeightFromGramsToKg(weightGrams) + " kg"
	} else {
		return formatWeightFromGramsToPounds(weightGrams) + " lbs"
	}
}

func formatHeightFeet(heightMm int) string {
	totalInches := float64(heightMm) / 25.4
	totalInchesRounded := int(totalInches + 0.5) // Round to nearest inch
	feetPart := totalInchesRounded / 12
	return fmt.Sprintf("%d", feetPart)
}

func formatHeightInches(heightMm int) string {
	totalInches := float64(heightMm) / 25.4
	totalInchesRounded := int(totalInches + 0.5) // Round to nearest inch
	inchesPart := totalInchesRounded % 12
	return fmt.Sprintf("%d", inchesPart)
}

func formatHeightCm(heightMm int) string {
	cm := float64(heightMm) / 10.0
	return fmt.Sprintf("%.0f", cm)
}

func formatWeightFromGramsToKg(weightGrams int) string {
	weightKg := float64(weightGrams) / 1000.0
	return fmt.Sprintf("%.1f", weightKg)
}

func formatWeightFromGramsToPounds(weightGrams int) string {
	weightKg := float64(weightGrams) / 1000.0
	pounds := weightKg * 2.20462
	return fmt.Sprintf("%.1f", pounds)
}

func formatHeightFromFeetInchesToCm(feet, inches int) int64 {
	totalInches := float64(feet)*12 + float64(inches)
	return int64(totalInches * 2.54)
}
