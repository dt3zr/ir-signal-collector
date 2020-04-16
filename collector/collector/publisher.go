package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func publish(frameRequest signalPublishRequest) {
	payload, err := json.Marshal(frameRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	response, err := http.Post("http://localhost:8080/signal", "application/json", bytes.NewReader(payload))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Published frame. Response %v received\n", response.StatusCode)
}
