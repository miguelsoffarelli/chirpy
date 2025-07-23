package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetApiKey(headers http.Header) (string, error) {
	apiKey := strings.Fields(headers.Get("Authorization"))
	// If the length of the resulting array is different than 2
	// or if the second element is empty, return an error
	if len(apiKey) != 2 || apiKey[1] == "" {
		return "", fmt.Errorf("invalid authentication header")
	}

	return apiKey[1], nil
}
