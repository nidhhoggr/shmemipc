package shmemipc

import (
	"fmt"
)

func StartServer(name string, len uint64, flags int) (*ShmProvider, error) {
	shm := ShmProvider{}

	err := shm.Listen(name, len, flags)
	if err != nil {
		fmt.Printf("Listen (%s)(%d) failed: %s", name, len, err.Error())
		return nil, err
	}

	return &shm, nil
}

func StartClient(name string) (*ShmProvider, error) {
	shm := ShmProvider{}

	err := shm.Dial(name)
	if err != nil {
		fmt.Println("Dial failed:" + err.Error())
		return nil, err
	}

	return &shm, nil
}

type BidirectionalShmem struct {
	shmResp *ShmProvider
	errResp error
	shmRqst *ShmProvider
	errRqst error
}

func (ir *BidirectionalShmem) Close() error {
	err := ir.shmResp.Close(nil)
	if err != nil {
		return err
	}
	return ir.shmRqst.Close(nil)
}

func (ir *BidirectionalShmem) GetError() error {
	if ir.errResp != nil {
		return ir.errResp
	}
	return ir.errRqst
}
