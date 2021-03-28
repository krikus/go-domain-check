package main

import (
	"container/list"
	"sync"
)

type ChanValidationResult struct {
	domain string
	result bool
}

// ChanBroker - broker instance
type ChanBroker struct {
	concurrency     int
	jobsChannel     chan string
	internalChan    chan *ChanValidationResult
	resultChan      chan *ChanValidationResult
	validator       func(string) bool
	queueOfRequests *list.List
	arrayOfResults  []*ChanValidationResult
	waitGroup       *sync.WaitGroup
	finished        chan bool
}

func CreateChanBroker(concurrency int, validatorFunc func(string) bool) *ChanBroker {
	arrayOfResults := make([]*ChanValidationResult, 0)
	finished := make(chan bool)
	internalChan := make(chan *ChanValidationResult)
	jobsChannel := make(chan string, concurrency)
	queueOfRequests := list.New()
	resultChan := make(chan *ChanValidationResult)

	broker := &ChanBroker{
		concurrency,
		jobsChannel,
		internalChan,
		resultChan,
		validatorFunc,
		queueOfRequests,
		arrayOfResults,
		&sync.WaitGroup{},
		finished,
	}

	broker.waitGroup.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go broker.spawnWorker()
	}

	go broker.spawnOverseerer()

	return broker
}

func (broker *ChanBroker) PushToValidate(domain string) {
	broker.queueOfRequests.PushBack(domain)
	broker.jobsChannel <- domain
}

func (broker *ChanBroker) GetResultsChan() chan *ChanValidationResult {
	return broker.resultChan
}

func (broker *ChanBroker) Close() {
	close(broker.jobsChannel)
	broker.internalChan <- &ChanValidationResult{domain: ""}
	<-broker.finished
}

func (broker *ChanBroker) Finished() {
	close(broker.finished)
}

// ----------------

func (broker *ChanBroker) tryToDequeue() bool {
	first := broker.queueOfRequests.Front()
	resultsNum := len(broker.arrayOfResults)

	if first != nil && resultsNum > 0 {
		for i := 0; i < resultsNum; i++ {

			result := broker.arrayOfResults[i]
			if result.domain == first.Value {
				broker.arrayOfResults = append(broker.arrayOfResults[:i], broker.arrayOfResults[i+1:]...)
				broker.queueOfRequests.Remove(first)
				broker.resultChan <- result

				return true
			}
		}
	}
	return false
}

func (broker *ChanBroker) spawnWorker() {
	defer (func() {
		broker.waitGroup.Done()
	})()

	for task := range broker.jobsChannel {
		broker.internalChan <- &ChanValidationResult{
			task,
			broker.validator(task),
		}
	}
}

func (broker *ChanBroker) spawnOverseerer() {
	defer (func() {
		close(broker.resultChan)
		if len(broker.arrayOfResults) > 0 {
			panic("[o] Please report this issue: #bfbne1")
		}
	})()

	for result := range broker.internalChan {
		if result.domain != "" {
			broker.arrayOfResults = append(broker.arrayOfResults, result)

			for broker.tryToDequeue() {
			}
		} else {
			go (func() {
				broker.waitGroup.Wait()
				close(broker.internalChan)
			})()
		}
	}
}
