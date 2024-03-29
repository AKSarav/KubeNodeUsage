package utils

import "strings"

var ValidColors = map[string]bool{
	"red":    true,
	"green":  true,
	"orange": true,
}

var ValidMetrics = map[string]bool{
	"memory": true,
	"disk":   true,
	"cpu":    true,
}

var ValidSorts = map[string]bool{
	"name":     true,
	"node":     true,
	"free":     true,
	"usage":    true,
	"color":    true,
	"capacity": true,
	"max":      true,
}

func IsValidColor(input string) bool {
	_, match := ValidColors[input]
	return match // if matched true else false
}

func IsValidSort(input string) bool {
	_, match := ValidSorts[input]
	return match // if matched true else false
}

func IsValidMetric(input string) bool {
	_, match := ValidMetrics[input]
	return match // if matched true else false
}

func PrintValidColors() []string {
	var result []string
	for k := range ValidColors {
		result = append(result, k)
	}
	return result
}

func PrintValidMetrics() string {
	var result []string
	for k := range ValidMetrics {
		result = append(result, k)
	}
	return "Choose one of ["+strings.Join(result, ", ")+"]"
}

func PrintValidSorts() string {
	var result []string
	for k := range ValidSorts {
		result = append(result, k)
	}
	// return comma separated string
	return "Choose one of ["+strings.Join(result, ", ")+"]"
}
