package shmemipc

import (
	"fmt"
	"testing"
	"time"
)

var TEST_COUNT = 500000

func Test(t *testing.T) {
	server, err := StartServer("test", 100)
	if err != nil {
		panic(err)
	}
	defer server.Close(nil)

	client, err := StartClient("test")
	if err != nil {
		panic(err)
	}
	defer client.Close(nil)

	go serverRoutine(server, client)

	clientRoutine(server, client)

	time.Sleep(time.Second * 1)
}

func log(i int, info string) {
	if i%10000 == 0 || i == TEST_COUNT-1 {
		fmt.Println(info)
	}
}

func serverRoutine(server *ShmProvider, client *ShmProvider) {

	for i := 0; i < TEST_COUNT; i++ {
		msg, err := server.Read()
		if err != nil {
			panic(err)
		}
		log(i, fmt.Sprintf("[server] [%d] Read from client: %s\n", i, string(msg)))
		clientMessage := "Hello, client!"
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
