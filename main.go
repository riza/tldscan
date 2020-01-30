package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	version = "v1.0"
)

var (
	domainFlag, fileNameListFlag, outputFileFlag string
	currentTLDs                                  []string
	outputFile                                   *os.File
	err                                          error
)

func init() {
	flag.StringVar(&domainFlag, "d", "", "Domain")
	flag.StringVar(&fileNameListFlag, "l", "", "TLD list json file (ex:  cctld.json,tld.json)")
	flag.StringVar(&outputFileFlag, "o", "out.txt", "Output file name")
}

func banner(totalTLD string) {

	var str = `                
| | | |   | |                    
| |_| | __| |___  ___ __ _ _ __  
| __| |/ _  / __|/ __/ _  | '_ \ 
| |_| | (_| \__ \ (_| (_| | | | |
 \__|_|\__,_|___/\___\__,_|_| |_| ` + version

	str += "\n\n Total TLDs to scan: " + totalTLD
	str += "\n-------------------------------------"

	fmt.Println(str)
}

func main() {
	flag.Parse()

	if len(domainFlag) <= 1 {
		log.Fatal("Domain must be not empty")
	}

	fileNames := strings.Split(fileNameListFlag, ",")

	for _, fileName := range fileNames {
		list, err := openList(fileName)

		if err != nil {
			log.Fatal(err)
		}

		currentTLDs = append(currentTLDs, list...)
	}

	outputFile, err = os.OpenFile(outputFileFlag, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalf("output file opening error: %s", err)
	}

	totalTLDs := strconv.FormatInt(int64(len(currentTLDs)), 10)

	banner(totalTLDs)

	start := time.Now()
	foundCount := 0

	var wg sync.WaitGroup
	wg.Add(len(currentTLDs))

	for _, tld := range currentTLDs {
		domain := domainFlag + "." + tld
		go func() {
			defer wg.Done()
			_, err := net.LookupCNAME(domain)
			if err == nil {
				foundCount++
				logFound(domain)
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)
	pwd, _ := os.Getwd()
	outputFilePath := pwd + "/" + outputFileFlag

	result := "Scan finished, scanner took %s, founded %d domain, output saved to %s\n"

	fmt.Printf(result, elapsed, foundCount, outputFilePath)

}

func logFound(domain string) {
	fmt.Printf("\033[1;32m[FOUND]\033[0m %s\n", domain)
	_, err := outputFile.WriteString(domain + "\n")
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
}

func openList(fileName string) (list []string, err error) {
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &list)

	if err != nil {
		return
	}

	return
}
