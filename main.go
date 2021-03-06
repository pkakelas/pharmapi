package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gocolly/colly/v2"
)

// Pharmacy represents what else, a pharmacy
type Pharmacy struct {
	ID           string `json:"id"`
	Brand        string `json:"brand"`
	Municipality string `json:"municipality"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Schedule     string `json:"schedule"`
	Open         bool   `json:"open"`
}

// Response is the response object of the JSON API
type Response struct {
	StatusCode int
	Pharmacies []Pharmacy
}

// URL is our scraping target
const URL = "https://fsa-efimeries.gr/"

func main() {
	addr := ":8080"

	http.HandleFunc("/", handler)
	log.Println("Magic is happening on", addr)

	log.Fatal(http.ListenAndServe(addr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request: " + getIP(r))
	response := &Response{Pharmacies: []Pharmacy{}}
	c := colly.NewCollector()

	c.OnHTML("#table", func(e *colly.HTMLElement) {
		var pharmacies = []Pharmacy{}
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			pharmacies = append(pharmacies, parsePharmacyRow(row))
		})

		response.Pharmacies = pharmacies
	})

	c.OnResponse(func(r *colly.Response) {
		response.StatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
		response.StatusCode = r.StatusCode
	})

	c.Visit(URL)

	// dump results
	json, err := json.Marshal(response)
	if err != nil {
		log.Println("Failed to serialize response:", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

func parsePharmacyRow(row *colly.HTMLElement) Pharmacy {
	open, _ := strconv.ParseBool(row.ChildText("td:nth-child(8)"))

	return Pharmacy{
		ID:           row.ChildText("td:nth-child(2)"),
		Municipality: row.ChildText("td:nth-child(3)"),
		Brand:        row.ChildText("td:nth-child(4)"),
		Address:      row.ChildText("td:nth-child(5)"),
		Phone:        row.ChildText("td:nth-child(6)"),
		Schedule:     row.ChildText("td:nth-child(7)"),
		Open:         open,
	}
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
