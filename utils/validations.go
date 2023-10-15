package utils

var ValidColors = map[string]bool{
	"red": true,
	"green": true,
	"orange": true,
}

var ValidMetrics = map[string]bool{
	"memory": true,
	"disk": true,
	"cpu": true,
}

var ValidSorts = map[string]bool{
	"name": true,
	"used": true,
	"color": true,
	"capacity": true,
}

func IsValidColor(input string) bool{
	_, match := ValidColors[input]
	return match // if matched true else false
}

func IsValidSort(input string) bool{
	_, match := ValidSorts[input]
	return match // if matched true else false
}

func IsValidMetric(input string) bool{
	_, match := ValidMetrics[input]
	return match // if matched true else false
}

func PrintValidColors()[]string{
	var result []string
	for k := range ValidColors{
		result = append(result, k)
	}
	return result
}

func PrintValidMetrics()[]string{
	var result []string
	for k := range ValidMetrics{
		result = append(result, k)
	}
	return result
}

func PrintValidSorts()[]string{
	var result []string
	for k := range ValidSorts{
		result = append(result, k)
	}
	return result
}