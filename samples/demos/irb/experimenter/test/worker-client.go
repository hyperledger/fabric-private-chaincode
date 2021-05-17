package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/protos"
)

func testGet() error {
	resp, err := http.Get("http://localhost:5000/proto-test")
	if err != nil {
		panic(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	serverSE := pb.ProtoTest{}
	err = proto.Unmarshal(bodyBytes, &serverSE)
	if err != nil {
		panic(err)
	}

	if serverSE.StudyId != "this is a study" || serverSE.Counter != 17 {
		return errors.New(fmt.Sprintf("unexpected return: %s - %d", serverSE.StudyId, serverSE.Counter))
	}

	return nil
}

func testPost() error {
	clientStudyId := "This is a client study"
	clientCounter := 20

	clientSE := pb.ProtoTest{}
	clientSE.StudyId = clientStudyId
	clientSE.Counter = int32(clientCounter)

	request, err := proto.Marshal(&clientSE)
	if err != nil {
		fmt.Println("respo1 err")
		panic(err)
	}

	resp, err := http.Post("http://localhost:5000/proto-test", "", bytes.NewBuffer(request))
	if err != nil {
		fmt.Println("respo 2err")
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("respo 3err")
		panic(err)
	}

	serverSE := pb.ProtoTest{}
	err = proto.Unmarshal(bodyBytes, &serverSE)
	if err != nil {
		return errors.New(fmt.Sprintf("unexpected return: %s", string(bodyBytes)))
	}

	if serverSE.StudyId != clientStudyId || serverSE.Counter != int32(clientCounter+1) {
		return errors.New(fmt.Sprintf("unexpected return: %s - %d", serverSE.StudyId, serverSE.Counter))
	}

	return nil
}

func main() {
	clientSE := pb.ProtoTest{}
	clientSE.StudyId = "this a client study"
	clientSE.Counter = 18

	_, err := proto.Marshal(&clientSE)
	if err != nil {
		panic(err)
	}

	fmt.Println("Testing get...")
	err = testGet()
	if err != nil {
		panic(err)
	}

	fmt.Println("Testing post...")
	err = testPost()
	if err != nil {
		panic(err)
	}

	fmt.Println("Test done.")
}
