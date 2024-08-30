package shmemipc

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/joe-at-startupmedia/go-arrow/arrow/memory"
)

const (
	INDEXOFFSET  = 0
	ENCLENOFFSET = 4
	DATAOFFSET   = 8
)

func (smp *ShmProvider) initEncoderDecoder() {
	// leave 4 bytes for the length of the message
	smp.buffer = *memory.NewBufferBytes(smp.ipcBuffer[DATAOFFSET:])
}

// Waits for messages or cancellation
// Writes a 1 to the index to indicate that the message has been read
func (smp *ShmProvider) Read() ([]byte, error) {

	// loop forever
	smp.bufmu.Lock()
	defer smp.bufmu.Unlock()

	for !smp.closed {

		// Wait for a message
		smp.bufmu.Unlock()
		smp.waitforevent(smp.wrevent)
		smp.bufmu.Lock()
		if smp.closed {
			break
		}

		// Were we woken up prematurely?
		index := binary.LittleEndian.Uint32(smp.ipcBuffer[INDEXOFFSET:])
		if index == 1 {
			continue
		}

		// Process the message
		encodingLen := binary.LittleEndian.Uint32(smp.ipcBuffer[ENCLENOFFSET:])
		smp.buffer.ResizeNoShrink(int(encodingLen))
		request := smp.buffer.Bytes()

		// Signal that we have read the data
		smp.signalevent(smp.rdevent)

		return request, nil
	}
	return nil, nil
}

func (smp *ShmProvider) ReadTimed(duration time.Duration) ([]byte, error) {

	readMsgChan := make(chan *[]byte, 1)
	readErrChan := make(chan error, 1)

	go func() {
		m, err := smp.Read()
		readMsgChan <- &m
		readErrChan <- err
	}()

	select {
	case <-time.After(duration):
		go func() {
			msg := <-readMsgChan
			if msg != nil {
				smp.Write(*msg)
			}
		}()
		return nil, errors.New("timed_out")
	case msg := <-readMsgChan:
		return *msg, <-readErrChan
	}
}

// Send function
// Writes a 1 to the index to indicate that the message has been written
func (smp *ShmProvider) Write(data []byte) error {

	smp.bufmu.Lock()
	defer smp.bufmu.Unlock()

	if smp.closed {
		return errors.New("buffer is closed")
	}

	binary.LittleEndian.PutUint32(smp.ipcBuffer[INDEXOFFSET:], uint32(0))
	encodingLen := uint32(len(data))
	binary.LittleEndian.PutUint32(smp.ipcBuffer[ENCLENOFFSET:], uint32(encodingLen))
	copy(smp.ipcBuffer[DATAOFFSET:], data)

	// Signal the reader and wait for response
	smp.signalevent(smp.wrevent)

	return nil
}
