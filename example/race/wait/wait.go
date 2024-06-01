package main

import (
	"github.com/joe-at-startupmedia/shmemipc"
	"github.com/joe-at-startupmedia/shmemipc/example"
	"sync"
	"time"
)

func main() {

	filename := "/tmp/example_simple_race_wait"
	wg := &sync.WaitGroup{}

	wg.Add(example.TEST_COUNT)

	server, err := shmemipc.StartServer(filename, 100, 0)
	if err != nil {
		panic(err)
	}
	defer server.Close(wg)

	client, err := shmemipc.StartClient(filename)
	if err != nil {
		panic(err)
	}
	defer client.Close(wg)

	go example.ServerRoutine(server, client)
	example.ClientRoutine(server, client, wg)

	time.Sleep(time.Second * 1)
}
