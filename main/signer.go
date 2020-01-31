package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

func main() {
	ExecutePipeline()
}

func ExecutePipeline(jobs ...job) {
	waitGroup := &sync.WaitGroup{}

	bufferSize := 2

	channel1 := createChannel(bufferSize)
	channel2 := createChannel(bufferSize)
	channel3 := createChannel(bufferSize)
	channel4 := createChannel(bufferSize)

	waitGroup.Add(1)
	go executePipeline(waitGroup, channel1, channel2, SingleHash)
	waitGroup.Add(1)
	go executePipeline(waitGroup, channel2, channel3, MultiHash)
	waitGroup.Add(1)
	go executePipeline(waitGroup, channel3, channel4, CombineResults)

	channel1 <- 0
	channel1 <- 1

	close(channel1)
	finalResult := <-channel4

	fmt.Println("CombineResults", finalResult)
	waitGroup.Wait()
}

func executePipeline(waitGroup *sync.WaitGroup, channel1 chan interface{}, channel2 chan interface{}, jobHash job) {
	defer waitGroup.Done()
	jobHash(channel1, channel2)
}

func SingleHash(in, out chan interface{}) {
	defer close(out)
	for data := range in {
		dataString := strconv.Itoa(data.(int))
		hash := DataSignerCrc32(dataString) + "~" + DataSignerCrc32(DataSignerMd5(dataString))
		out <- hash
	}

}

func MultiHash(in, out chan interface{}) {
	defer close(out)
	for data := range in {
		dataString := data.(string)
		buffer := make([]string, 6)
		for i := 0; i < 5; i++ {
			formatInt := strconv.Itoa(i)
			number := DataSignerCrc32(formatInt + dataString)
			buffer = append(buffer, number)
		}
		finalResult := ""
		for idx := range buffer {
			finalResult += buffer[idx]
		}
		out <- finalResult
	}
}

func CombineResults(in, out chan interface{}) {
	defer close(out)
	data := make([]string, 0)
	for value := range in {
		data = append(data, value.(string))
	}
	sort.Strings(data)
	finalResult := ""
	length := len(data)
	for idx := range data {
		if idx+1 == length {
			finalResult += data[idx]
		} else {
			finalResult += data[idx] + "_"
		}
	}
	out <- finalResult
}

func createChannel(bufferSize int) chan interface{} {
	return make(chan interface{})
}
