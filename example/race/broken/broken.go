package main

import (
	"github.com/joe-at-startupmedia/shmemipc"
	"github.com/joe-at-startupmedia/shmemipc/example"
	"time"
)

// main because both the server and client will be reading
// from the same memory, this will behave non-deterministically
// whereby the client might read its own data before the
// server has a chance to. This can lead to multiple scenarios:
//
//  1. The server never reads because the client always reads
//     the data first
//  2. The server reads a few times and the client reads a few
//     of its messages faster than the server
//  3. The client and server both read and write to each other
//     as expected
//
// In scenarios 1 and 2: the server will never close because it
// will stay open waiting for to read messages that will never
// be sent. if the closing of the server is handled by using defer
// in the main block, it will cause a panic because the server is
// still attempting to read.
//
// race_fixed.go
//
// These problems can be solved by using channels to prevent the
// server/client from reading their own writes or a duplex
// connection can be used instead as provided in the duplex example.
//
// race_wait.go
//
// To prevent the server from being closed while it is
// still reading, a wait-group can be leveraged and passed to the
// close function see race_wait.go in this package
func main() {

	filename := "/tmp/example_simple_race"

	server, err := shmemipc.StartServer(filename, 100)
	if err != nil {
		panic(err)
	}

	client, err := shmemipc.StartClient(filename)
	if err != nil {
		panic(err)
	}
	defer client.Close(nil)

	go func() {
		example.ServerRoutine(server, client)
		//uncommenting these will result in the server reading its own writes
		//serverRoutine(server, client)
		//serverRoutine(server, client)
		//we have to place close here to prevent panics
		server.Close(nil)
	}()

	example.ClientRoutine(server, client, nil)

	time.Sleep(time.Second * 1)
}
