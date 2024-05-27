package shmemipc

import "time"

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

func (ir *IpcRequester) ReadTimed(duration time.Duration) ([]byte, error) {
	return ir.shmResp.ReadTimed(duration)
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
