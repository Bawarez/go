package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
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
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, url := range urls {
		for i:= 0; i < requestsNumber; i++ {
			wg.Add(1)
			go func() {
				channel := make(chan int64)
				go trace(url, channel)
				select {
				case ping := <- channel:
					mutex.Lock()
					responseTime = append(responseTime, ping)
					mutex.Unlock()
				case <- time.After(time.Duration(int(time.Millisecond) * timeout)):
					mutex.Lock()
					failedRequests++
					mutex.Unlock()
				}
				defer wg.Done()
			}()
		}
	}

	wg.Wait()
	printResults(responseTime, failedRequests)
}

func trace(url string, channel chan int64)   {
	start := time.Now()
	_, err := http.Get(url)
	if err != nil {
		return
	}
	duration := time.Now().Sub(start)

	//convert to milliseconds
	channel <- duration.Nanoseconds() / 1000000
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
	for i := 1; i < len(slice); i++ {
		if slice[i] > max {
			max = slice[i]
		} else if slice[i] < min  {
			min = slice[i]
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
	length := int64(len(slice))
	if length == 0 {
		return 0
	}
	return sum(slice) / length
}
