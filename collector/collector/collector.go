package collector

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/tarm/serial"
)

func parseRequiredFlags() (string, int) {
	serialPort := flag.String("serial", "", "Specifies the serial port in the form /dev/xxx")
	baudRate := flag.Int("baud", 9600, "Specifies the baud rate of the serial port")

	flag.Parse()

	if len(*serialPort) < 1 {
		fmt.Println("Serial port not specified")
		flag.Usage()
	}
	if _, err := os.Stat(*serialPort); err != nil {
		log.Fatal(err)
	}

	return *serialPort, *baudRate
}

// Start launches the collector
func Start() {
	serialPort, baudRate := parseRequiredFlags()
	config := &serial.Config{Name: serialPort, Baud: baudRate}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	lineScanner := bufio.NewScanner(io.Reader(port))
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, os.Interrupt)

		fmt.Println("Press Ctrl-C to exit program")

		<-signalChannel

		log.Println("Closing serial port")
		port.Close()
		os.Exit(0)
	}()
	type TaggedFrame struct {
		CollectorID string           `json:"collectorId,omitempty"`
		Frame       *json.RawMessage `json:"frame"`
	}
	for lineScanner.Scan() {
		frameJSON := lineScanner.Bytes()
		log.Printf("Received frame -> %s", string(frameJSON))
		var taggedFrame TaggedFrame
		if err := json.Unmarshal(frameJSON, &taggedFrame); err != nil {
			log.Println(err)
		}
		taggedFrame.CollectorID = "giant"
		if taggedFrameJSON, err := json.Marshal(taggedFrame); err != nil {
			log.Println(err)
		} else {
			go publishTaggedFrameJSON(taggedFrameJSON)
		}
	}
	if err := lineScanner.Err(); err != nil {
		log.Println(err)
	}
}
