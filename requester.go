package shmemipc

import (
	"bytes"
	"fmt"
	"io"
)

type IpcRequester struct {
	BidirectionalShmem
	conn *IpcRequesterConn
}

func NewRequester(filename string) *IpcRequester {
	responder, errResp := StartClient(filename + "_resp")
	requester, errRqst := StartClient(filename + "_rqst")

	ir := IpcRequester{
		BidirectionalShmem: BidirectionalShmem{
			shmResp: responder,
			errResp: errResp,
			shmRqst: requester,
			errRqst: errRqst,
		},
	}

	ir.conn = &IpcRequesterConn{
		requester: &ir,
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
	return ir.BidirectionalShmem.GetError()
}

func (ir *IpcRequester) Close() error {
	return ir.BidirectionalShmem.Close()
}

type IpcRequesterConn struct {
	requester *IpcRequester
}

func (ir *IpcRequester) GetConn() *IpcRequesterConn {
	return ir.conn
}

func (ir *IpcRequesterConn) Read(b []byte) (int, error) {
	msg, err := ir.requester.Read()
	if err != nil {
		return 0, err
	}
	fmt.Printf("[requester_conn] read msg: %s %s\n", string(msg), b)
	src := bytes.NewBuffer(msg)
	dst := bytes.NewBuffer(b)
	_, err = io.Copy(dst, src)
	fmt.Printf("[requester_conn] read src: %s, dest: %s, msg: %s\n", src, dst, string(msg))
	return 0, err
}

func (ir *IpcRequesterConn) Write(b []byte) (int, error) {
	fmt.Printf("[requester_conn] write msg: %s\n", string(b))
	err := ir.requester.Write(b)
	return 0, err
}

func (ir *IpcRequesterConn) Close() error {
	return ir.requester.Close()
}
