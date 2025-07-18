package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	// Two different types: one for OK response and the other for errors
	type returnValsOK struct {
		Valid bool `json:"valid"`
	}

	type returnValsErr struct {
		Error string `json:"error"`
	}

	// this one is defined in the scope of the whole function because it'll be used more than once
	errBody := returnValsErr{
		Error: "Something went wrong",
	}

	errDat, err := json.Marshal(errBody)
	if err != nil {
		log.Printf("Error marshaling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// decode request body
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		writeResponse(http.StatusBadRequest, errDat, w)
		return
	}

	// length check
	if len(params.Body) > 140 {
		tooLongBody := returnValsErr{
			Error: "Chirp is too long",
		}

		tooLongDat, err := json.Marshal(tooLongBody)
		if err != nil {
			writeResponse(http.StatusInternalServerError, errDat, w)
			return
		}

		writeResponse(http.StatusBadRequest, tooLongDat, w)
		return
	}

	// OK response
	respBody := returnValsOK{
		Valid: true,
	}

	OKDat, err := json.Marshal(respBody)
	if err != nil {
		writeResponse(http.StatusInternalServerError, errDat, w)
		return
	}

	writeResponse(http.StatusOK, OKDat, w)
}

func writeResponse(code int, data []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}
