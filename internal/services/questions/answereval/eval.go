package answereval

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	epsilon = 1e-3 // Tolerance for decimal comparison
)

// EvaluateAnswer compares a student's answer against the correct answer
// It handles multiple formats: integers, decimals, and fractions
func EvaluateAnswer(studentAnswer string, correctAnswer string) (bool, error) {
	// Parse both answers
	studentVal, err := parseAnswer(strings.TrimSpace(studentAnswer))
	if err != nil {
		return false, fmt.Errorf("invalid student answer: %v", err)
	}

	correctVal, err := parseAnswer(strings.TrimSpace(correctAnswer))
	if err != nil {
		return false, fmt.Errorf("invalid correct answer: %v", err)
	}

	// Compare with epsilon tolerance
	return math.Abs(studentVal-correctVal) < epsilon, nil
}

// parseAnswer converts various answer formats to a float64
// Supports: integers (9), decimals (3.5), fractions (18/2), negative numbers (-5)
func parseAnswer(answer string) (float64, error) {
	answer = strings.TrimSpace(answer)

	// Handle empty answer
	if answer == "" {
		return 0, fmt.Errorf("empty answer")
	}

	// Check if it's a fraction (contains /)
	if strings.Contains(answer, "/") {
		return parseFraction(answer)
	}

	// Otherwise parse as decimal/integer
	val, err := strconv.ParseFloat(answer, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %s", answer)
	}

	return val, nil
}

// parseFraction converts fraction strings like "18/2", "1/3", "-7/2" to float64
func parseFraction(fraction string) (float64, error) {
	parts := strings.Split(strings.TrimSpace(fraction), "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid fraction format: %s", fraction)
	}

	numerator, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numerator: %s", parts[0])
	}

	denominator, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid denominator: %s", parts[1])
	}

	if denominator == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	return numerator / denominator, nil
}

// EvaluateAnswerVerbose provides detailed feedback on the comparison
type EvaluationResult struct {
	IsCorrect       bool
	StudentValue    float64
	CorrectValue    float64
	Difference      float64
	StudentFormatted string
	CorrectFormatted string
}

// EvaluateAnswerWithFeedback compares answers and returns detailed feedback
func EvaluateAnswerWithFeedback(studentAnswer string, correctAnswer string) (*EvaluationResult, error) {
	studentVal, err := parseAnswer(strings.TrimSpace(studentAnswer))
	if err != nil {
		return nil, fmt.Errorf("invalid student answer: %v", err)
	}

	correctVal, err := parseAnswer(strings.TrimSpace(correctAnswer))
	if err != nil {
		return nil, fmt.Errorf("invalid correct answer: %v", err)
	}

	diff := math.Abs(studentVal - correctVal)
	isCorrect := diff < epsilon

	return &EvaluationResult{
		IsCorrect:        isCorrect,
		StudentValue:     studentVal,
		CorrectValue:     correctVal,
		Difference:       diff,
		StudentFormatted: formatDecimal(studentVal),
		CorrectFormatted: formatDecimal(correctVal),
	}, nil
}

func formatDecimal(val float64) string {
	if val == math.Floor(val) {
		return fmt.Sprintf("%.0f", val)
	}
	formatted := fmt.Sprintf("%.4f", val)
	formatted = strings.TrimRight(formatted, "0")
	return strings.TrimSuffix(formatted, ".") 
}

