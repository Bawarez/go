package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func main() {
	var urls string
	var requestsNumber int
	var timeout int

	flag.StringVar(&urls, "u", "", "user urls")
	flag.IntVar(&requestsNumber, "rn", 1, "requests number")
	flag.IntVar(&timeout, "t", 500, "timeout")
	flag.Parse()

	traceAll(strings.Split(urls, " "),requestsNumber, timeout)
}

func traceAll(urls []string, requestsNumber int, timeout int, ) {
	responseTime := make([]int64, 0)
	failedRequests := 0

	for _, url := range urls {
		for i:= 0; i < requestsNumber; i++ {
			channel := make(chan int64, 1)
			go func() {
				ping, err := trace(url)
				if err == nil {
					channel <- ping
				}
			}()
			select {
			case ping := <- channel:
				responseTime = append(responseTime, ping)
			case <- time.After(time.Duration(int(time.Millisecond) * timeout)):
				failedRequests++
			}
		}
	}

	printResults(responseTime, failedRequests)
}

func trace(url string) (int64, error) {
	start := time.Now()
	_, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	duration := time.Now().Sub(start)

	//convert to milliseconds
	return duration.Nanoseconds() / 1000000, nil
}

func printResults(responseTime []int64, failedRequests int) {
	fmt.Println("Total response time:", sum(responseTime))
	fmt.Println("Average response time:", avg(responseTime))

	min, max, err := getMinAndMax(responseTime)
	if err == nil {
		fmt.Println("max: ", max)
		fmt.Println("min: ", min)
	}

	fmt.Println("Failed requests number: ", failedRequests)
}

func getMinAndMax(slice []int64) (int64, int64, error) {
	if len(slice) == 0 {
		return 0, 0, errors.New("Empty slice")
	}

	max := slice[0]
	min := slice[0]
	for _, element := range slice {
		if element > max {
			max = element
		} else if element < min  {
			min = element
		}
	}
	return min, max, nil
}

func sum(slice []int64) int64 {
	var sum int64 = 0
	for _, element := range slice {
		sum += element
	}
	return sum
}

func avg(slice []int64) int64 {
	return sum(slice) / int64(len(slice))
}