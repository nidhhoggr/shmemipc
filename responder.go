package shmemipc

import "time"

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

func (ir *IpcResponder) ReadTimed(duration time.Duration) ([]byte, error) {
	return ir.shmRqst.ReadTimed(duration)
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
