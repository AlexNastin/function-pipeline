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
	waitGroupSingleHash := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	for data := range in {
		waitGroupSingleHash.Add(1)
		go asyncIterationSingleHash(mutex, data, out, waitGroupSingleHash)
	}
	waitGroupSingleHash.Wait()
}

func asyncIterationSingleHash(mutex *sync.Mutex, data interface{}, out chan interface{}, waitGroupSingleHash *sync.WaitGroup) {
	defer waitGroupSingleHash.Done()
	dataString := strconv.Itoa(data.(int))
	md5 := hashMD5(mutex, dataString)
	firstCrc32 := hashCrc32(dataString)
	secondCrc32 := hashCrc32(<-md5)
	hash := <-firstCrc32 + "~" + <-secondCrc32
	out <- hash
}

func MultiHash(in, out chan interface{}) {
	waitGroupMultiHash := &sync.WaitGroup{}
	for data := range in {
		waitGroupMultiHash.Add(1)
		go asyncIterationMultiHash(data, out, waitGroupMultiHash)
	}
	waitGroupMultiHash.Wait()
}

func asyncIterationMultiHash(data interface{}, out chan interface{}, waitGroupMultiHash *sync.WaitGroup) {
	defer waitGroupMultiHash.Done()
	dataString := data.(string)
	buffer := make([]string, 6)
	for i := 0; i <= 5; i++ {
		formatInt := strconv.Itoa(i)
		number := hashCrc32(formatInt + dataString)
		buffer = append(buffer, <-number)
	}
	finalResult := ""
	for idx := range buffer {
		finalResult += buffer[idx]
	}
	out <- finalResult
}

func CombineResults(in, out chan interface{}) {
	//waitGroupCombineResults := &sync.WaitGroup{}
	data := make([]string, 0)
	for value := range in {
		//waitGroupCombineResults.Add(1)
		//go asyncIterationCombineResults(data, value, waitGroupCombineResults)
		data = append(data, value.(string))
	}
	//waitGroupCombineResults.Wait()
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

func asyncIterationCombineResults(data []string, value interface{}, waitGroupCombineResults *sync.WaitGroup) {
	defer waitGroupCombineResults.Done()
	data = append(data, value.(string))
}

func hashMD5(mutex *sync.Mutex, dataString string) chan string {
	result := make(chan string, 1)
	go func(out chan<- string) {
		mutex.Lock()
		md5 := DataSignerMd5(dataString)
		mutex.Unlock()
		out <- md5
	}(result)
	return result
}

func hashCrc32(dataString string) chan string {
	result := make(chan string, 1)
	go func(out chan<- string) {
		crc32 := DataSignerCrc32(dataString)
		out <- crc32
	}(result)
	return result
}

func createChannel() chan interface{} {
	return make(chan interface{})
}
