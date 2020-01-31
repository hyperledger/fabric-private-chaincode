/*

Copyright Intel Corp. 2019 All Rights Reserved.
SPDX-License-Identifier: Apache-2.0

*/

/* Notes

   Simple utility which reads the payload for a request from a file and submits the corresponding
   request to the fabric backend. Naming conventions determine the file to be picked with a format
   '<scenario-dir>/<caller>.<action>[.<round>].json'.

   For usage, run it with '--help', by default it creates an auction in the localhost backend

*/

/* TODO
 - add json files in demo/scenario for a complete scenario
 - make both backends use demo/scenario/Auctioneer1.createAuction.json instead of their own
   (see also below for generalization)
   - mock-server might potentially re-use below loadRequestPayload
     (which then might have to be outsourced in separate utility file?)

- maybe:
  - revisit logging? use shimlogger as server.go does? Doesn't really seem the right way, though?
  - extend backends & UI to also offer these files as a Get similar to the auction
    template as a quick-fill to UI (i.e., generalize /getDefaultAuction)
    - e.g., a '/getDefault' with action and (optional) round as (json?) request params
      and user via x-user
    - scenario-path would be provided by cmd-line arg of gw
  - smarter return processing? (e.g., status.rc as exit code?  auctionId, if existing, prominent?)
  - implement variant which calls directly peer
    func fpcPeerRequest(requestName string, user string, payloadObj interface{}) {}
*/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type debugging bool

var dlog debugging

func (d debugging) Printf(format string, args ...interface{}) {
	if dlog {
		log.Printf(format, args...)
	}
}

// note: round and auctionId are optional and considered absent if negative
func loadRequestPayload(scenarioPath string, user string, requestName string, round int, auctionId int) (interface{}, error) {

	// check whether scenario directory exists
	// - first check right away ...
	fInfo, err := os.Stat(scenarioPath)
	// - if it doesn't exist, try also by prefixing FPC_PATH env-var if existing
	if err != nil || !fInfo.IsDir() {
		fpcPath, exists := os.LookupEnv("FPC_PATH")
		if exists {
			scenarioPath = fmt.Sprintf("%s%c%s", fpcPath, os.PathSeparator, scenarioPath)
			fInfo, err = os.Stat(scenarioPath)
		}
	}
	if err != nil || !fInfo.IsDir() {
		return nil, fmt.Errorf("illegal scenario path '%s'", scenarioPath)
	}

	// try to open requestPayload file
	var requestPayloadPath string
	if round >= 0 {
		requestPayloadPath = fmt.Sprintf("%s%c%s.%s.%d.json", scenarioPath, os.PathSeparator, user, requestName, round)
	} else {
		requestPayloadPath = fmt.Sprintf("%s%c%s.%s.json", scenarioPath, os.PathSeparator, user, requestName)
	}
	jsonFile, err := os.Open(requestPayloadPath)
	defer jsonFile.Close()
	if err != nil {
		return nil, err
	}

	// read requestPayload  ...
	byteRequestPayload, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	// decode requestPayload  ...
	var jsonRequestPayload interface{}
	err = json.Unmarshal(byteRequestPayload, &jsonRequestPayload)
	if err != nil {
		return nil, err
	}

	// replace auctionId if requested
	if auctionId >= 0 {
		switch jsonRequestPayload.(type) {
		case map[string]interface{}:
			obj := jsonRequestPayload.(map[string]interface{})
			if _, ok := obj["auctionId"]; ok {
				dlog.Printf("Replacing auctionId '%v' with '%v'", obj["auctionId"], auctionId)
				obj["auctionId"] = auctionId
			}
		}
	}

	// return object ...
	return jsonRequestPayload, nil
}

func fpcBackendRequest(url string, requestName string, requestPayloadObj interface{}, user string) error {

	requestPayloadBytes, _ := json.Marshal(requestPayloadObj)
	requestPayloadStr := string(requestPayloadBytes)

	type FpcBackendRequest struct {
		Tx   string   `json:"tx"`
		Args []string `json:"args"`
	}
	request := FpcBackendRequest{
		Args: []string{requestPayloadStr},
		Tx:   requestName,
	}
	requestBytes, _ := json.Marshal(request)
	requestStr := string(requestBytes)

	dlog.Printf("serialized request payload: '%s'\n\n", requestPayloadStr)
	dlog.Printf("serialized request: '%s'\n\n", requestStr)

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-user", user)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Successful requestion\n- http-status=%d\n- http-body='%s'\n", resp.StatusCode, body)
	return nil

}

func main() {
	scenarioPathPtr := flag.String("scenario-path", "demo/scenario", "path to directory containing requests for choosen scenario (searched directly and prefixed with FPC_PATH env var, iff existing)")
	requestNamePtr := flag.String("request", "createAuction", "name of request")
	roundPtr := flag.Int("round", -1, "round number, for request with multiple calls for same user (optional, must be >= 0)")
	auctionIdPtr := flag.Int("auction-id", -1, "id used to replace 'auctionId' properties in payload (optional, must be >= 0)")
	userPtr := flag.String("user", "Auctioneer1", "user initiating request")

	urlPtr := flag.String("url", "http://localhost:3000/api/cc/invoke", "URL to use for fpc backend call")

	dryRunPtr := flag.Bool("dry-run", false, "show request on stdout but do not execute it")

	debugPtr := flag.Bool("debug", false, "debug invocation")

	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Printf("Illegal trailing parameters %s\nUsage of %s:\n", flag.Args(), os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *debugPtr {
		dlog = true
	} else {
		dlog = false
	}

	requestPayloadObj, err := loadRequestPayload(*scenarioPathPtr, *userPtr, *requestNamePtr, *roundPtr, *auctionIdPtr)
	if err != nil {
		log.Fatalf("Could not read requestPayload object: err='%s'", err)
	}

	if *dryRunPtr {
		prettyJSON, _ := json.MarshalIndent(requestPayloadObj, "", "    ")
		fmt.Printf("%s\n", prettyJSON)
	} else {

		err = fpcBackendRequest(*urlPtr, *requestNamePtr, requestPayloadObj, *userPtr)
		if err != nil {
			log.Fatalf("Could not submit request '%s' with payload '%v': err='%s'", *requestNamePtr, requestPayloadObj, err)
		}
	}

}
