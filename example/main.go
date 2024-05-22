package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	shmemipc "github.com/joe-at-startupmedia/shmemipc"
)

var msgIndex int

// onNewMessage is the callback function that is called when a new message is received
func onNewMessage(data []byte, requestMetadata map[string]string) ([]byte, int, string) {
	fmt.Printf("[server] [%d] Read from client: %s\n", msgIndex, string(data))
	clientMessage := "Hello, client!"
	fmt.Printf("[server] [%d] Write to client: %s, 200, OK\n\n", msgIndex, clientMessage)
	msgIndex++
	return []byte(clientMessage), 200, "OK"
}

func serverRoutine(ctx context.Context, shm *shmemipc.ShmProvider) {

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		shm.Receive(ctx, onNewMessage)
		wg.Done()
	}()

	//shm.Close(&wg)
}

func clientRoutine(ctx context.Context, shm *shmemipc.ShmProvider) {
	metadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	serverMessage := "Hello, server!"
	fmt.Println("[client] Write to server: " + serverMessage)

	// Send the message to the server
	response, status, statusMessage := shm.Send(ctx, []byte(serverMessage), metadata)
	fmt.Println("[client] Response from server: " + string(response) + ", " + strconv.Itoa(int(status)) + ", " + statusMessage)

	//shm.Close(nil)
}

// main is the main function
func main() {
	// create a shared memory provider
	ctx := context.Background()
	shm := shmemipc.ShmProvider{}
	err := shm.Listen(ctx, "/tmp/shmipc")
	if err != nil {
		// this is the server because the shared memory
		// does not exist yet
		err := shm.Dial(ctx, "/tmp/shmipc", 100)
		if err != nil {
			fmt.Println("Dial failed:" + err.Error())
			return
		}
	}

	// write to or read from shared memory

	go serverRoutine(ctx, &shm)

	time.Sleep(time.Second * 4)

	clientRoutine(ctx, &shm)
	shm.Close(nil)
}
