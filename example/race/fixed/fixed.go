package main

import (
	"fmt"
	"github.com/joe-at-startupmedia/shmemipc"
	"github.com/joe-at-startupmedia/shmemipc/example"
	"time"
)

var clientReadChannel = make(chan bool, example.TEST_COUNT)
var serverReadChannel = make(chan bool, example.TEST_COUNT)

func ServerRoutine(server *shmemipc.ShmProvider, client *shmemipc.ShmProvider) {
	clientMessage := "Hello, client! "
	for i := 0; i < example.TEST_COUNT; i++ {
		msg, err := server.Read()
		if err != nil {
			panic(err)
		}
		example.Log(i, fmt.Sprintf("[server] [%d] Read from client: %s\n", i, string(msg)))
		serverReadChannel <- true
		err = client.Write([]byte(fmt.Sprintf("%s %d", clientMessage, i)))
		if err != nil {
			panic(err)
		}
		example.Log(i, fmt.Sprintf("[server] [%d] Write to client: %s\n", i, clientMessage))
		<-clientReadChannel
	}
}

func ClientRoutine(server *shmemipc.ShmProvider, client *shmemipc.ShmProvider) {

	serverMessage := "Hello, server! "

	for i := 0; i < example.TEST_COUNT; i++ {
		example.Log(i, fmt.Sprintf("[client] [%d] Write to server: %s\n", i, serverMessage))
		err := server.Write([]byte(fmt.Sprintf("%s %d", serverMessage, i)))
		if err != nil {
			panic(err)
		}
		<-serverReadChannel
		msg, err := client.Read()
		if err != nil {
			panic(err)
		}
		example.Log(i, fmt.Sprintf("[client] [%d] Response from server: %s\n", i, string(msg)))
		clientReadChannel <- true
	}
}

// main while the following solution fixes the race conditions
// in race.go, it is also impractical since the server and client
// run in separate application which don't have access to
// each-others channels. As such a similar inter-process
// communication methods must be utilized (i.e. os signals)
func main() {

	filename := "/tmp/example_simple"

	server, err := shmemipc.StartServer(filename, 100, 0)
	if err != nil {
		panic(err)
	}
	defer server.Close(nil)

	client, err := shmemipc.StartClient(filename)
	if err != nil {
		panic(err)
	}
	defer client.Close(nil)

	go ServerRoutine(server, client)
	ClientRoutine(server, client)

	time.Sleep(time.Second * 1)
}
