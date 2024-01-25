package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("shakesearch available at http://localhost:%s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
			queryValues := r.URL.Query()
			query, ok := queryValues["q"]
			if !ok || len(query[0]) < 1 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("missing search query in URL params"))
					return
			}

			// Initialize default values for pagination.
			page := 0
			pageSize := 20
			// Parse 'page' parameter from URL query, if present.
			if p, ok := queryValues["page"]; ok {
					parsedPage, err := strconv.Atoi(p[0])
					if err == nil {
							page = parsedPage
					}
			}
			// Parse 'pageSize' parameter from URL query, if present.
			if ps, ok := queryValues["pageSize"]; ok {
					parsedPageSize, err := strconv.Atoi(ps[0])
					if err == nil && parsedPageSize > 0 {
							pageSize = parsedPageSize
					}
			}

			results := searcher.Search(query[0], page, pageSize)
			buf := &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			err := enc.Encode(results)
			if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("encoding failure"))
					return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
			return fmt.Errorf("Load: %w", err)
	}
	// Convert the "complete works" text to lowercase to ensure case-insensitive search.
	lowerCaseText := strings.ToLower(string(dat))
	s.CompleteWorks = lowerCaseText
	s.SuffixArray = suffixarray.New([]byte(lowerCaseText))
	return nil
}

func (s *Searcher) Search(query string, page, pageSize int) []string {
	// Convert the query to lowercase to ensure case-insensitive search.
	lowerCaseQuery := strings.ToLower(query)
	idxs := s.SuffixArray.Lookup([]byte(lowerCaseQuery), -1)
	
	// Calculate start and end index for slicing the results
	startIdx := page * pageSize
	endIdx := startIdx + pageSize
	if endIdx > len(idxs) {
			endIdx = len(idxs)
	}

	var results []string
	for _, idx := range idxs[startIdx:endIdx] {
			start := max(0, idx-250)
			end := min(len(s.CompleteWorks), idx+250)
			results = append(results, s.CompleteWorks[start:end])
	}
	return results
}

// max returns the larger of two integers, used for ensuring slice start index is not negative.
func max(a, b int) int {
	if a > b {
			return a
	}
	return b
}

// min returns the smaller of two integers, used for limiting a slice's end index.
func min(a, b int) int {
	if a < b {
			return a
	}
	return b
}
