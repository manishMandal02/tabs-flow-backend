package utils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"

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

func MakeHTTPRequest(method, url string, headers map[string]string, body []byte) (*http.Response, string, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, "", err
	}

	return resp, string(respBody), nil
}
