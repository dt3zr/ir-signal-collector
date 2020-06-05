package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func BenchmarkDatabaseInsert(b *testing.B) {
	// data := "{\"collectorId\":\"id\", \"frame\":{\"resolution\":20,\"data\":[]}}"
	// server := httptest.NewServer(http.HandlerFunc(frameQueryHandler))
	// defer server.Close()
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			data := []byte(`{"collectorId":"arduinoble","frame":{"resolution":20,"data":[[452,226],[103,28],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,27],[29,84],[29,84],[30,27],[30,84],[29,84],[30,83],[30,84],[30,83],[30,27],[30,27],[29,28],[30,83],[31,27],[29,28],[29,28],[29,27],[31,83],[31,83],[29,84],[29,28],[30,83],[29,85],[29,84],[29,84]]}}`)
			reader := bytes.NewReader([]byte(data))
			// _, err := http.Post(server.URL, "application/json", reader)
			// if err != nil {
			// 	b.Error(err)
			// }
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/ir/frame", reader)
			frameQueryHandler(w, r)
		}()
		time.Sleep(20 * time.Nanosecond)
	}
	wg.Wait()
}
