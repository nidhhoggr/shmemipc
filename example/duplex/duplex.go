package main

import (
	"fmt"
	"github.com/joe-at-startupmedia/shmemipc"
	"time"
)

var TEST_COUNT = 5

func log(i int, info string) {
	fmt.Println(info)
}

func responderRoutine(ir *shmemipc.IpcResponder) {
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

func requesterRoutine(ir *shmemipc.IpcRequester) {

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

// main is the main function
func main() {
	filename := "/tmp/example_duplex"
	responder := shmemipc.NewResponder(filename, 100, 0)
	if err := responder.GetError(); err != nil {
		panic(err)
	}

	requester := shmemipc.NewRequester(filename)
	if err := requester.GetError(); err != nil {
		panic(err)
	}
	defer requester.Close()

	go func() {
		responderRoutine(responder)
		//uncommenting these will result in the responders hanging
		//responderRoutine(responder)
		//responderRoutine(responder)
		responder.Close()
	}()

	requesterRoutine(requester)

	time.Sleep(time.Second * 1)
}
