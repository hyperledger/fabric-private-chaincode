package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/data-provider/service/pdf"
)

func main() {

	inputFile := flag.String("input", "", "path to file to parse")
	flag.Parse()

	pdfContent, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		log.Fatal(err)
	}

	patientInfo, err := pdf.ParseQuestionForm(pdfContent)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("contents -> %s\n", patientInfo)
	fmt.Printf("patient answers -> %s\n", patientInfo.Answers.ToString())
}
