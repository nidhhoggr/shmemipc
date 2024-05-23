package shmemipc

import "fmt"

func StartServer(name string, len uint64) (*ShmProvider, error) {
	shm := ShmProvider{}

	err := shm.Listen(name, len)
	if err != nil {
		fmt.Println("Listen failed:" + err.Error())
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

type IpcResponder BidirectionalShmem

func NewResponder(filename string, len uint64) *IpcResponder {
	responder, errResp := StartServer(filename+"_resp", len)
	requester, errRqst := StartServer(filename+"_rqst", len)

	ir := IpcResponder{
		shmResp: responder,
		errResp: errResp,
		shmRqst: requester,
		errRqst: errRqst,
	}

	return &ir
}

func (ir *IpcResponder) Read() ([]byte, error) {
	return ir.shmRqst.Read()
}

func (ir *IpcResponder) Write(b []byte) error {
	return ir.shmResp.Write(b)
}

func (ir *IpcResponder) GetError() error {
	return (*BidirectionalShmem)(ir).GetError()
}

func (ir *IpcResponder) Close() error {
	return (*BidirectionalShmem)(ir).Close()
}

type IpcRequester BidirectionalShmem

func NewRequester(filename string) *IpcRequester {
	responder, errResp := StartClient(filename + "_resp")
	requester, errRqst := StartClient(filename + "_rqst")

	ir := IpcRequester{
		shmResp: responder,
		errResp: errResp,
		shmRqst: requester,
		errRqst: errRqst,
	}

	return &ir
}

func (ir *IpcRequester) Read() ([]byte, error) {
	return ir.shmResp.Read()
}

func (ir *IpcRequester) Write(b []byte) error {
	return ir.shmRqst.Write(b)
}

func (ir *IpcRequester) GetError() error {
	return (*BidirectionalShmem)(ir).GetError()
}

func (ir *IpcRequester) Close() error {
	return (*BidirectionalShmem)(ir).Close()
}
