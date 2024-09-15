package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

	"github.com/google/uuid"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

// Generates UUID v7, fallback to UUID v4 if errored while generating V7
func GenerateID() string {
	id, err := uuid.NewV7()
	if err != nil {
		logger.Error("Error generating UUID v7", err)
		return uuid.NewString()
	}

	return id.String()
}

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func GenerateOTP() string {
	maxDigits := 6
	bi, err := rand.Int(
		rand.Reader,
		big.NewInt(int64(math.Pow(10, float64(maxDigits)))),
	)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%0*d", maxDigits, bi)
}
