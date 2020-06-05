package server

import (
	"bytes"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func BenchmarkDatabaseInsertion(b *testing.B) {
	initDb := func(done <-chan interface{}) (<-chan frameCRUD, chan<- frameCRUD) {
		db := newDatabase()
		dbLock := make(chan frameCRUD)
		dbUnlock := make(chan frameCRUD)
		go func() {
			defer close(dbLock)
			defer close(dbUnlock)
			dbLock <- db
			for {
				select {
				case ref := <-dbUnlock:
					dbLock <- ref
				case <-done:
					return
				}
			}
		}()
		return dbLock, dbUnlock
	}
	done := make(chan interface{})
	defer close(done)
	dbLock, dbUnlock = initDb(done)

	wList := make([]http.ResponseWriter, b.N)
	rList := make([]*http.Request, b.N)
	for i := 0; i < b.N; i++ {
		data := []byte(`{"collectorId":"arduinoble","frame":{"resolution":20,"data":[[452,226],[103,28],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,27],[29,84],[29,84],[30,27],[30,84],[29,84],[30,83],[30,84],[30,83],[30,27],[30,27],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,83],[31,83],[29,84],[29,28],[30,83],[29,85],[29,84],[29,84]]}}`)
		reader := bytes.NewReader([]byte(data))
		rList[i] = httptest.NewRequest(http.MethodPost, "http://localhost:8080/ir/frame", reader)
		wList[i] = httptest.NewRecorder()
	}

	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func(i int) {
			defer wg.Done()
			frameQueryHandler(wList[i], rList[i])
		}(i)
	}
	wg.Wait()
}

func BenchmarkDatabaseQueryCollector(b *testing.B) {
	initDb := func(done <-chan interface{}) (<-chan frameCRUD, chan<- frameCRUD) {
		db := newDatabase()
		dbLock := make(chan frameCRUD)
		dbUnlock := make(chan frameCRUD)
		go func() {
			defer close(dbLock)
			defer close(dbUnlock)
			dbLock <- db
			for {
				select {
				case ref := <-dbUnlock:
					dbLock <- ref
				case <-done:
					return
				}
			}
		}()
		return dbLock, dbUnlock
	}
	done := make(chan interface{})
	defer close(done)
	dbLock, dbUnlock = initDb(done)

	data := []byte(`{"collectorId":"arduinoble","frame":{"resolution":20,"data":[[452,226],[103,28],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,27],[29,84],[29,84],[30,27],[30,84],[29,84],[30,83],[30,84],[30,83],[30,27],[30,27],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,83],[31,83],[29,84],[29,28],[30,83],[29,85],[29,84],[29,84]]}}`)
	reader := bytes.NewReader([]byte(data))
	r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/ir/frame", reader)
	w := httptest.NewRecorder()
	frameQueryHandler(w, r)

	wList := make([]http.ResponseWriter, b.N)
	rList := make([]*http.Request, b.N)
	for i := 0; i < b.N; i++ {
		rList[i] = httptest.NewRequest(http.MethodGet, "http://localhost:8080/ir/collector", nil)
		wList[i] = httptest.NewRecorder()
	}

	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func(i int) {
			defer wg.Done()
			collectorQueryHandler(wList[i], rList[i])
		}(i)
	}
	wg.Wait()
}

func BenchmarkDatabaseMixed(b *testing.B) {
	initDb := func(done <-chan interface{}) (<-chan frameCRUD, chan<- frameCRUD) {
		db := newDatabase()
		dbLock := make(chan frameCRUD)
		dbUnlock := make(chan frameCRUD)
		go func() {
			defer close(dbLock)
			defer close(dbUnlock)
			dbLock <- db
			for {
				select {
				case ref := <-dbUnlock:
					dbLock <- ref
				case <-done:
					return
				}
			}
		}()
		return dbLock, dbUnlock
	}
	done := make(chan interface{})
	defer close(done)
	dbLock, dbUnlock = initDb(done)

	wList := make([]http.ResponseWriter, b.N)
	rList := make([]*http.Request, b.N)
	writeN := int(b.N / 2)
	readN := b.N - writeN
	for i := 0; i < writeN; i++ {
		data := []byte(`{"collectorId":"arduinoble","frame":{"resolution":20,"data":[[452,226],[103,28],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,27],[29,84],[29,84],[30,27],[30,84],[29,84],[30,83],[30,84],[30,83],[30,27],[30,27],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,83],[31,83],[29,84],[29,28],[30,83],[29,85],[29,84],[29,84]]}}`)
		reader := bytes.NewReader([]byte(data))
		rList[i] = httptest.NewRequest(http.MethodPost, "http://localhost:8080/ir/frame", reader)
		wList[i] = httptest.NewRecorder()
	}
	for i := writeN; i < (writeN + readN); i++ {
		rList[i] = httptest.NewRequest(http.MethodGet, "http://localhost:8080/ir/collector", nil)
		wList[i] = httptest.NewRecorder()
	}
	p := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		p[i] = i
	}
	rand.Shuffle(len(p), func(i, j int) {
		p[i], p[j] = p[j], p[i]
	})

	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for _, i := range p {
		go func(i int) {
			defer wg.Done()
			if rList[i].Method == http.MethodPost {
				frameQueryHandler(wList[i], rList[i])
			} else {
				collectorQueryHandler(wList[i], rList[i])
			}
		}(i)
	}
	wg.Wait()
}
