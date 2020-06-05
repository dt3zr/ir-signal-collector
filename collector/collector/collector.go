package collector

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/tarm/serial"
)

type flagSet struct {
	serialPort *string
	baudRate   *int
	serverHost *string
	serverPort *int
	cid        *string
}

func (fs *flagSet) parseRequiredFlags() error {
	fs.serialPort = flag.String("serial", "", "Specifies the serial port in the form /dev/xxx")
	fs.baudRate = flag.Int("baud", 9600, "Specifies the baud rate of the serial port")
	fs.serverHost = flag.String("server", "localhost", "Specifies host name or IP address or server")
	fs.serverPort = flag.Int("port", 8080, "Specifies the port number of the server")
	fs.cid = flag.String("collectorId", "", "Specifies the id of this instance of collector")
	flag.Parse()
	if len(*fs.cid) < 1 {
		flag.Usage()
		return errors.New("Collector ID not specified")
	}
	if len(*fs.serialPort) < 1 {
		flag.Usage()
		return errors.New("Serial port not specified")
	}
	return nil
}

// Start launches the collector
func Start() {
	var collectorFlag flagSet
	collectorFlag.parseRequiredFlags()

	if _, err := os.Stat(*collectorFlag.serialPort); err != nil {
		log.Fatal("Error checking serial port.", err)
	}
	config := &serial.Config{Name: *collectorFlag.serialPort, Baud: *collectorFlag.baudRate}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal("Error opening serial port.", err)
	}
	defer port.Close()
	log.Printf("Opened serial port '%s' at baud rate %d", *collectorFlag.serialPort, *collectorFlag.baudRate)
	lineScanner := bufio.NewScanner(io.Reader(port))
	client, err := newPublishClient(*collectorFlag.serverHost, *collectorFlag.serverPort)
	if err != nil {
		log.Fatal("Error creating publish client.", err)
	}
	log.Printf("Frames will be published to '%s'", client.serverURL)
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, os.Interrupt)

		log.Print("Press Ctrl-C to exit program")

		<-signalChannel

		log.Print("Closing serial port")
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
		taggedFrame.CollectorID = *collectorFlag.cid
		if taggedFrameJSON, err := json.Marshal(taggedFrame); err != nil {
			log.Println("Error marshaling.", err)
		} else {
			go client.publishTaggedFrameJSON(taggedFrameJSON)
		}
	}
	if err := lineScanner.Err(); err != nil {
		log.Println(err)
	}
}
