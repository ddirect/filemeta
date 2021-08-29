package filemeta

import (
	"runtime"
	"sync"
)

type Async struct {
	FileIn  chan string
	DataOut chan Data
}

func AsyncOperations(op Op, probeThreads int, hashThreads int) Async {
	if op == OpRefresh {
		return AsyncMono(op)
	}
	if probeThreads < 1 {
		probeThreads = runtime.NumCPU()
	}
	if hashThreads < 1 {
		hashThreads = probeThreads
	}
	bufSize := probeThreads * 500
	fileIn := make(chan string, bufSize)
	dataOut := make(chan Data, bufSize)
	hashingIn := make(chan Data, bufSize)

	var wg1, wg2 sync.WaitGroup
	wg1.Add(probeThreads)
	wg2.Add(hashThreads)

	for i := 0; i < probeThreads; i++ {
		go func() {
			defer wg1.Done()
			for file := range fileIn {
				data := core(op, file)
				if data.hashNeeded {
					hashingIn <- data
				} else {
					dataOut <- data
				}
			}
		}()
	}

	for i := 0; i < hashThreads; i++ {
		go func() {
			defer wg2.Done()
			h := getHasher()
			defer h.done()
			for data := range hashingIn {
				data.notifyHash(h.run(data.Path, data.Size))
				dataOut <- data
			}
		}()
	}

	go func() {
		wg1.Wait()
		close(hashingIn)
		wg2.Wait()
		close(dataOut)
	}()
	return Async{fileIn, dataOut}
}

func AsyncMono(op Op) Async {
	bufSize := 100
	fileIn := make(chan string, bufSize)
	dataOut := make(chan Data, bufSize)
	go func() {
		h := getHasher()
		defer h.done()
		for file := range fileIn {
			data := core(op, file)
			if data.hashNeeded {
				data.notifyHash(h.run(file, data.Size))
			}
			dataOut <- data
		}
		close(dataOut)
	}()
	return Async{fileIn, dataOut}
}
