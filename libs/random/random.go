package random

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	mathrand "math/rand"
	"time"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// -------------------- integer methods  -------------------

// RandRangeInt64 Choose a random int64 from range [min, max]
func RandRangeInt64(min, max int64) int64 {
	if min > max {
		logger.Panic("The min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))
		return result.Int64() - i64Min
	}
	result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return min + result.Int64()
}

// RandRangeInt Choose a random int from range [min, max]
func RandRangeInt(min, max int) int {
	return int(RandRangeInt64(int64(min), int64(max)))
}

// -------------------- sequence methods  -------------------

// Choice .
func Choice(seq []interface{}) interface{} {
	if len(seq) == 0 {
		return seq
	}
	return seq[RandRangeInt(0, len(seq)-1)]
}

// Shuffle .
func Shuffle(x []interface{}) {
	r := mathrand.New(mathrand.NewSource(time.Now().Unix()))
	for len(x) > 0 {
		n := len(x)
		randIndex := r.Intn(n)
		x[n-1], x[randIndex] = x[randIndex], x[n-1]
		x = x[:n-1]
	}
}

// Sample Chooses k unique random elements from a population sequence or set
func Sample(population []interface{}, k int) []interface{} {
	Shuffle(population)
	min := RandRangeInt(0, len(population)-k)
	return population[min : min+k]
}

// Choices Return a k sized list of population elements chosen with replacement
// TODO
func Choices(population []interface{}, weights float64, k int) []interface{} {
	Shuffle(population)
	min := RandRangeInt(0, len(population)-k)
	return population[min : min+k]
}

// -------------------- use case  -------------------

// ChoiceStrArr Choice for []string
func ChoiceStrArr(arr []string) string {
	s := make([]interface{}, len(arr))
	for i, v := range arr {
		s[i] = v
	}
	return fmt.Sprintf("%v", Choice(s))
}

// SampleStrArr Choice for []string
func SampleStrArr(arr []string, k int) []string {
	s := make([]interface{}, len(arr))
	for i, v := range arr {
		s[i] = v
	}
	spArr := make([]string, k)
	for j, sp := range Sample(s, k) {
		spArr[j] = fmt.Sprintf("%v", sp)
	}
	return spArr
}
