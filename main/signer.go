package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

func main() {
	inputData := []int{0, 1}
	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			defer close(out)
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, _ := dataRaw.(string)
			testResult := data
			fmt.Print(testResult)
		}),
	}

	ExecutePipeline(hashSignJobs...)
}

func ExecutePipeline(jobs ...job) {
	waitGroup := &sync.WaitGroup{}

	channels := make([]chan interface{}, len(jobs)+1)

	for i := 0; i < len(jobs)+1; i++ {
		channels[i] = createChannel()
	}

	for idx, job := range jobs {
		waitGroup.Add(1)
		go executePipeline(waitGroup, channels[idx], channels[idx+1], job)
	}

	waitGroup.Wait()
}

func executePipeline(waitGroup *sync.WaitGroup, in chan interface{}, out chan interface{}, jobHash job) {
	defer waitGroup.Done()
	defer close(out)
	jobHash(in, out)
}

func SingleHash(in, out chan interface{}) {
	for data := range in {
		dataString := strconv.Itoa(data.(int))
		hash := DataSignerCrc32(dataString) + "~" + DataSignerCrc32(DataSignerMd5(dataString))
		out <- hash
	}
}

func MultiHash(in, out chan interface{}) {
	for data := range in {
		dataString := data.(string)
		buffer := make([]string, 6)
		for i := 0; i <= 5; i++ {
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

func createChannel() chan interface{} {
	return make(chan interface{})
}
