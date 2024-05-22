package main

import (
	"fmt"
	shmemipc "github.com/joe-at-startupmedia/shmemipc"
	"sync"
	"time"
)

var msgIndex int

// onNewMessage is the callback function that is called when a new message is received
func newMessage(data []byte) string {
	fmt.Printf("[server] [%d] Read from client: %s\n", msgIndex, string(data))
	clientMessage := "Hello, client!"
	fmt.Printf("[server] [%d] Write to client: %s\n", msgIndex, clientMessage)
	msgIndex++
	return clientMessage
}

func serverRoutine(server *shmemipc.ShmProvider, client *shmemipc.ShmProvider) {

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		msg, err := server.Read()
		if err != nil {
			panic(err)
		}
		err = client.Write([]byte(newMessage(msg)))
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	wg.Wait()
}

func clientRoutine(server *shmemipc.ShmProvider, client *shmemipc.ShmProvider) {

	serverMessage := "Hello, server!"
	fmt.Println("[client] Write to server: " + serverMessage)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		err := server.Write([]byte(serverMessage))
		if err != nil {
			panic(err)
		}
		msg, err := client.Read()
		if err != nil {
			panic(err)
		}
		fmt.Printf("[client] Response from server: %s\n", string(msg))
		wg.Done()
	}()

	wg.Wait()
}

// main is the main function
func main() {

	server, err := shmemipc.StartServer("example", 100)
	if err != nil {
		panic(err)
	}
	defer server.Close(nil)

	client, err := shmemipc.StartClient("example")
	if err != nil {
		panic(err)
	}
	defer client.Close(nil)

	go func() {
		serverRoutine(server, client)
		server.Close(nil)
	}()

	clientRoutine(server, client)

	time.Sleep(time.Second * 1)
}
