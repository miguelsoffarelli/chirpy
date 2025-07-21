package main

import (
	"strings"
)

func validateChirp(params *chirpParameters) bool {
	// length check
	const maxChirpLen = 140
	if len(params.Body) > maxChirpLen {
		return false
	}

	clearChirp(params)
	return true
}

func clearChirp(p *chirpParameters) {
	body := strings.Split(p.Body, " ")
	for i, word := range body {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			body[i] = "****"
		}
	}

	p.Body = strings.Join(body, " ")
}
