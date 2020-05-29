package collector

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
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

const (
	pNECHeaderMarkMicros float64 = 9000
	pNECBitShortMicros   float64 = 562.5
	pNECBitLongMicros    float64 = 1687.5
)

type irBit []int

type irFrameData struct {
	Resolution int     `json:"resolution"`
	Data       []irBit `json:"data"`
}

func matchNECProtocol(h markSpacePair) bool {
	return (float64(h.Mark) > pNECHeaderMarkMicros*0.90) && (float64(h.Mark) < pNECHeaderMarkMicros*1.1)
}

func (f *irFrameData) getProtocol() string {
	if matchNECProtocol(f.getHeader()) {
		return "NEC"
	}
	return "Unknown"
}

func (f *irFrameData) getPulseLength() int {
	return len(f.Data[1:])
}

func (f *irFrameData) getRawPulses() []markSpacePair {
	pulses := f.Data[1:]
	pair := make([]markSpacePair, len(pulses))

	for i, p := range pulses {
		pair[i].Mark = p[0] * f.Resolution
		pair[i].Space = p[1] * f.Resolution
	}

	return pair
}

func (f *irFrameData) getHeader() markSpacePair {
	return markSpacePair{
		f.Data[0][0] * f.Resolution,
		f.Data[0][1] * f.Resolution,
	}
}

func matchNECShort(microTime int) bool {
	return (float64(microTime) > pNECBitShortMicros*0.90) && (float64(microTime) < pNECBitShortMicros*1.1)
}

func matchNECLong(microTime int) bool {
	return (float64(microTime) > pNECBitLongMicros*0.90) && (float64(microTime) < pNECBitLongMicros*1.1)
}

func (f *irFrameData) getDecodedValue() (string, error) {
	var decodedValue uint32 = 0
	pulses := f.Data[1:33]

	for i, p := range pulses {
		markMicroTime := p[0] * f.Resolution
		spaceMicroTime := p[1] * f.Resolution

		isMarkShort := matchNECShort(markMicroTime)
		isSpaceShort := matchNECShort(spaceMicroTime)
		isSpaceLong := matchNECLong(spaceMicroTime)

		if isMarkShort && isSpaceLong {
			decodedValue = decodedValue + uint32(math.Pow(float64(2), float64(31-i)))
		} else if isMarkShort && !isSpaceShort {
			return "", errors.New("Error decoding value")
		}
	}

	return fmt.Sprintf("%08X", decodedValue), nil

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

		port.Close()
		os.Exit(0)
	}()

	for lineScanner.Scan() {

		line := lineScanner.Bytes()
		fmt.Println(string(line))

		var data irFrameData
		if err := json.Unmarshal(line, &data); err != nil {
			fmt.Println(err)
		}

		// packaging irFrameData into signalPublishRequest
		var request signalPublishRequest
		request.ProtocolName = data.getProtocol()
		request.FrameSize = data.getPulseLength()
		request.Value, err = data.getDecodedValue()
		request.Header = data.getHeader()
		request.RawPulses = data.getRawPulses()

		if err != nil {
			fmt.Println(err)
		} else {
			go publish(request)
		}

	}

	if err := lineScanner.Err(); err != nil {
		fmt.Println(err)
	}

}
