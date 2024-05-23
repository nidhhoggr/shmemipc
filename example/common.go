package example

import (
	"fmt"
	"github.com/joe-at-startupmedia/shmemipc"
	"sync"
)

var TEST_COUNT = 5

func Log(i int, info string) {
	fmt.Println(info)
}

func ServerRoutine(server *shmemipc.ShmProvider, client *shmemipc.ShmProvider) {
	clientMessage := "Hello, client!"
	for i := 0; i < TEST_COUNT; i++ {
		msg, err := server.Read()
		if err != nil {
			panic(err)
		}
		Log(i, fmt.Sprintf("[server] [%d] Read from client: %s\n", i, string(msg)))
		err = client.Write([]byte(clientMessage))
		if err != nil {
			panic(err)
		}
		Log(i, fmt.Sprintf("[server] [%d] Write to client: %s\n", i, clientMessage))
	}
}

func ClientRoutine(server *shmemipc.ShmProvider, client *shmemipc.ShmProvider, wg *sync.WaitGroup) {

	serverMessage := "Hello, server!"

	for i := 0; i < TEST_COUNT; i++ {
		Log(i, fmt.Sprintf("[client] [%d] Write to server: %s\n", i, serverMessage))
		err := server.Write([]byte(serverMessage))
		if err != nil {
			panic(err)
		}
		msg, err := client.Read()
		if err != nil {
			panic(err)
		}
		Log(i, fmt.Sprintf("[client] [%d] Response from server: %s\n", i, string(msg)))

		if wg != nil {
			wg.Done()
		}
	}
}
