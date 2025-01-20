package utils

import (
	"Planning_poker/app/models"
	"math"
	"strconv"
)

func CalculateVotingAverage(participants map[string]models.Participant) float64 {
	totalVotes := 0
	count := 0

	for _, participant := range participants {
		i, err := strconv.Atoi(participant.Vote)
		if err != nil {
			continue
		}

		if i == 0 {
			continue
		}

		totalVotes += i
		count++
	}

	if count == 0 {
		return 0
	}

	average := float64(totalVotes) / float64(count)
	roundedAverage := math.Round(average*10) / 10
	return roundedAverage
}

func CalculateFibonacciVotingAverage(participants map[string]models.Participant) int {
	fibonacci := []int{1, 2, 3, 5, 8, 13}
	totalVotes := 0
	count := 0

	for _, participant := range participants {
		i, err := strconv.Atoi(participant.Vote)
		if err != nil {
			continue
		}

		if isFibonacci(i, fibonacci) {
			totalVotes += i
			count++
		}
	}

	if count == 0 {
		return 0
	}

	average := float64(totalVotes) / float64(count)
	return roundToNearestFibonacci(average, fibonacci)
}

var tshirtSizes = map[string]int{
	"XS":  1,
	"S":   2,
	"M":   3,
	"L":   4,
	"XL":  5,
	"XXL": 6,
}

func CalculateTshirtsVotingAverage(participants map[string]models.Participant) string {
	totalVotes := 0
	totalWeight := 0

	for _, participant := range participants {
		if voteWeight, exists := tshirtSizes[participant.Vote]; exists {
			totalVotes += voteWeight
			totalWeight++
		}
	}

	if totalWeight == 0 {
		return "?"
	}

	average := float64(totalVotes) / float64(totalWeight)
	return roundToNearestTshirtSize(average)
}

func roundToNearestTshirtSize(average float64) string {
	nearest := "?"
	minDiff := float64(1000)

	for size, weight := range tshirtSizes {
		diff := math.Abs(float64(weight) - average)
		if diff < minDiff {
			minDiff = diff
			nearest = size
		}
	}
	return nearest
}

func isFibonacci(num int, fibonacci []int) bool {
	for _, f := range fibonacci {
		if num == f {
			return true
		}
	}
	return false
}

func roundToNearestFibonacci(num float64, fibonacci []int) int {
	for i := 1; i < len(fibonacci); i++ {
		low := fibonacci[i-1]
		high := fibonacci[i]

		midpoint := float64(low) + float64(high-low)/2

		if num <= midpoint {
			return low
		} else if num <= float64(high) {
			return high
		}
	}

	return fibonacci[len(fibonacci)-1]
}
