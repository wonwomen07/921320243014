package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"
)

type NumberResponse struct {
	Numbers []int `json:"numbers"`
}

func fetchNumbersFromURL(url string, ch chan<- []int) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data from", url, err)
		ch <- nil
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body from", url, err)
		ch <- nil
		return
	}

	var numResponse NumberResponse
	err = json.Unmarshal(body, &numResponse)
	if err != nil {
		fmt.Println("Error unmarshaling response from", url, err)
		ch <- nil
		return
	}

	ch <- numResponse.Numbers
}

func mergeUniqueNumbers(numbersList ...[]int) []int {
	uniqueNumbers := make(map[int]bool)
	merged := []int{}

	for _, numbers := range numbersList {
		for _, num := range numbers {
			if !uniqueNumbers[num] {
				uniqueNumbers[num] = true
				merged = append(merged, num)
			}
		}
	}

	sort.Ints(merged) // Use the sort package to sort the merged numbers

	return merged
}

func getMergedNumbersFromURLs(urls []string) []int {
	var wg sync.WaitGroup
	ch := make(chan []int, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			fetchNumbersFromURL(u, ch)
		}(url)
	}

	wg.Wait()
	close(ch)

	var numbersList [][]int
	for nums := range ch {
		if nums != nil {
			numbersList = append(numbersList, nums)
		}
	}

	return mergeUniqueNumbers(numbersList...)
}

func NumbersHandler(w http.ResponseWriter, r *http.Request) {
	urls, ok := r.URL.Query()["url"]
	if !ok {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	var validURLs []string
	for _, url := range urls {
		_, err := http.NewRequest(http.MethodGet, url, nil)
		if err == nil {
			validURLs = append(validURLs, url)
		}
	}

	startTime := time.Now()
	mergedNumbers := getMergedNumbersFromURLs(validURLs)
	endTime := time.Now()

	response := NumberResponse{
		Numbers: mergedNumbers,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
	fmt.Println("Time taken:", endTime.Sub(startTime))
}

func main() {
	http.HandleFunc("/numbers", NumbersHandler)

	port := ":3000"
	fmt.Println("Server is listening on port", port)
	http.ListenAndServe(port, nil)
}
