package shmemipc

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type IpcResponder struct {
	BidirectionalShmem
	conn *IpcResponderConn
}

func NewResponder(filename string, len uint64) *IpcResponder {
	responder, errResp := StartServer(filename+"_resp", len)
	requester, errRqst := StartServer(filename+"_rqst", len)

	ir := IpcResponder{
		BidirectionalShmem: BidirectionalShmem{
			shmResp: responder,
			errResp: errResp,
			shmRqst: requester,
			errRqst: errRqst,
		},
	}

	ir.conn = &IpcResponderConn{
		responder: &ir,
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
	return ir.BidirectionalShmem.GetError()
}

func (ir *IpcResponder) Close() error {
	return ir.BidirectionalShmem.Close()
}

func (ir *IpcResponder) Accept() (net.Conn, error) {
	panic("unimplemented")
}

func (ir *IpcResponder) Network() string {
	return "shmemipc"
}

func (ir *IpcResponder) String() string {
	return ir.shmResp.name
}

func (ir *IpcResponder) Addr() net.Addr {
	return ir
}

type IpcResponderConn struct {
	responder *IpcResponder
}

func (ir *IpcResponder) GetConn() *IpcResponderConn {
	return ir.conn
}

func (ir *IpcResponderConn) Read(b []byte) (int, error) {
	msg, err := ir.responder.Read()
	if err != nil {
		return 0, err
	}
	fmt.Printf("[responder_conn] read msg: %s %s\n", b, string(msg))
	src := bytes.NewBuffer(msg)
	dst := bytes.NewBuffer(b)
	_, err = io.Copy(dst, src)
	fmt.Printf("[responder_conn] read src: %s, dest: %s, msg: %s\n", src, dst, msg)
	return 0, err
}

func (ir *IpcResponderConn) Write(b []byte) (int, error) {
	fmt.Printf("[responder_conn] write msg: %s\n", string(b))
	err := ir.responder.Write(b)
	return 0, err
}

func (ir *IpcResponderConn) Close() error {
	return ir.responder.Close()
}
