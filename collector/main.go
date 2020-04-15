package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

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

func main() {

	serialPort, baudRate := parseRequiredFlags()
	config := &serial.Config{Name: serialPort, Baud: baudRate}
	port, err := serial.OpenPort(config)

	if err != nil {
		log.Fatal(err)
	}

	lineScanner := bufio.NewScanner(io.Reader(port))

	headers := make([]string, 0)
	records := make([][]string, 0)
	fields := make([]string, 0)

	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, os.Interrupt)

		fmt.Println("Press Ctrl-C to exit program")

		<-signalChannel

		fmt.Println("Saving data")

		file, err := os.Create("output.csv")
		if err != nil {
			log.Fatal(err)
		}

		outputWriter := csv.NewWriter(file)
		err = outputWriter.Write(headers)

		if err != nil {
			log.Fatal(err)
		}

		err = outputWriter.WriteAll(records)

		if err != nil {
			log.Fatal(err)
		}

		outputWriter.Flush()
		file.Close()
		port.Close()
		os.Exit(0)
	}()

	headersCreated := false
	messageHeaderSent := false

	for lineScanner.Scan() {

		line := lineScanner.Text()
		fmt.Println(line)

		if strings.HasPrefix(line, "Decoded") {

			valueScanner := bufio.NewScanner(strings.NewReader(line))
			valueScanner.Split(bufio.ScanWords)
			for valueScanner.Scan() {
				token := valueScanner.Text()
				if strings.HasPrefix(token, "Value") {
					dataValues := strings.Split(token, ":")
					fields = append(fields, dataValues[1])
					if !headersCreated {
						headers = append(headers, "value")
					}
				}
			}
			messageHeaderSent = true

		} else if strings.Contains(line, "Head") {

			leftTrimmedLine := strings.TrimLeft(line, " \t")
			token := strings.Split(leftTrimmedLine, " ")
			if !headersCreated {
				headers = append(headers, "hm")
				headers = append(headers, "hs")
			}
			fields = append(fields, strings.TrimLeft(token[1], "m"))
			fields = append(fields, strings.TrimLeft(token[3], "s"))

		} else if messageHeaderSent && (strings.HasPrefix(line, "0:") ||
			strings.HasPrefix(line, "4:") ||
			strings.HasPrefix(line, "8:") ||
			strings.HasPrefix(line, "12:") ||
			strings.HasPrefix(line, "16:") ||
			strings.HasPrefix(line, "20:") ||
			strings.HasPrefix(line, "24:") ||
			strings.HasPrefix(line, "28:")) {

			valueScanner := bufio.NewScanner(strings.NewReader(line))

			valueScanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				trimmedData := bytes.TrimLeft(data, " \t")
				sepIndex := bytes.IndexAny(trimmedData, ":\t")
				if sepIndex > -1 {
					return sepIndex + 1 + (len(data) - len(trimmedData)), trimmedData[:sepIndex], nil
				}
				return 0, nil, nil
			})

			for valueScanner.Scan() {
				token := valueScanner.Text()
				token = strings.TrimSpace(token)
				if _, err := strconv.ParseInt(token, 10, 64); err == nil {
					if !headersCreated {
						headers = append(headers, strings.Join([]string{token, "m"}, ""))
						headers = append(headers, strings.Join([]string{token, "s"}, ""))
					}
				} else {
					timingTokens := strings.Split(token, " ")
					fields = append(fields, strings.TrimLeft(timingTokens[0], "ms"))
					fields = append(fields, strings.TrimLeft(timingTokens[1], "ms"))
				}
			}

			if strings.HasPrefix(line, "28:") {
				headersCreated = true
				messageHeaderSent = false
				records = append(records, fields)
				fields = make([]string, 0)
				fmt.Println("Recorded", len(records), "frames")
			}

		}

	}

	if err := lineScanner.Err(); err != nil {
		fmt.Println(err)
	}

}
