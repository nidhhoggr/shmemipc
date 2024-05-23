package shmemipc

import (
	"fmt"
	"testing"
	"time"
)

var TEST_COUNT = 500000

func log(i int, info string) {
	if i%100000 == 0 || i == TEST_COUNT-1 {
		//if i == TEST_COUNT-1 {
		fmt.Println(info)
	}
}

func TestSimple(t *testing.T) {
	filename := getFilename("test_simple")
	server, err := StartServer(filename, 100)
	if err != nil {
		panic(err)
	}
	defer server.Close(nil)

	client, err := StartClient(filename)
	if err != nil {
		panic(err)
	}
	defer client.Close(nil)

	go serverRoutine(server, client)

	clientRoutine(server, client)

	time.Sleep(time.Second * 1)
}

func serverRoutine(server *ShmProvider, client *ShmProvider) {
	clientMessage := "Hello, client!"
	for i := 0; i < TEST_COUNT; i++ {
		msg, err := server.Read()
		if err != nil {
			panic(err)
		}
		log(i, fmt.Sprintf("[server] [%d] Read from client: %s\n", i, string(msg)))
		err = client.Write([]byte(clientMessage))
		if err != nil {
			panic(err)
		}
		log(i, fmt.Sprintf("[server] [%d] Write to client: %s\n", i, clientMessage))
	}
}

func clientRoutine(server *ShmProvider, client *ShmProvider) {

	serverMessage := "Hello, server!"

	for i := 0; i < TEST_COUNT; i++ {
		log(i, fmt.Sprintf("[client] [%d] Write to server: %s\n", i, serverMessage))
		err := server.Write([]byte(serverMessage))
		if err != nil {
			panic(err)
		}
		msg, err := client.Read()
		if err != nil {
			panic(err)
		}
		log(i, fmt.Sprintf("[client] [%d] Response from server: %s\n", i, string(msg)))
	}
}

// main is the main function
func TestDuplex(t *testing.T) {
	filename := getFilename("test_duplex")
	responder := NewResponder(filename, 100)
	if err := responder.GetError(); err != nil {
		panic(err)
	}
	defer responder.Close()

	requester := NewRequester(filename)
	if err := requester.GetError(); err != nil {
		panic(err)
	}
	defer requester.Close()

	go responderRoutine(responder)

	requesterRoutine(requester)

	time.Sleep(time.Second * 1)
}

func responderRoutine(ir *IpcResponder) {
	clientMessage := "Hello Requester"
	for i := 0; i < TEST_COUNT; i++ {
		msg, err := ir.Read()
		if err != nil {
			panic(err)
		}
		log(i, fmt.Sprintf("[responder] [%d] Read from requester: %s\n", i, string(msg)))
		err = ir.Write([]byte(clientMessage))
		if err != nil {
			panic(err)
		}
		log(i, fmt.Sprintf("[responder] [%d] Write to requester: %s\n", i, clientMessage))
	}
}

func requesterRoutine(ir *IpcRequester) {

	serverMessage := "Hello, responder!"

	for i := 0; i < TEST_COUNT; i++ {
		log(i, fmt.Sprintf("[requester] [%d] Write to server: %s\n", i, serverMessage))
		err := ir.Write([]byte(serverMessage))
		if err != nil {
			panic(err)
		}
		msg, err := ir.Read()
		if err != nil {
			panic(err)
		}
		log(i, fmt.Sprintf("[requester] [%d] Response from server: %s\n", i, string(msg)))
	}
}
