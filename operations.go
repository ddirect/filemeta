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
		hashThreads = 1
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
		run, done := SyncOperations(op)
		defer done()
		for file := range fileIn {
			dataOut <- run(file)
		}
		close(dataOut)
	}()
	return Async{fileIn, dataOut}
}

func SyncOperations(op Op) (func(string) Data, func()) {
	var h *hasher
	return func(file string) (data Data) {
			data = core(op, file)
			if data.hashNeeded {
				if h == nil {
					h = getHasher()
				}
				data.notifyHash(h.run(file, data.Size))
			}
			return
		}, func() {
			if h != nil {
				h.done()
				h = nil
			}
		}
}
