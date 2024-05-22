package shmemipc

import (
	"encoding/binary"
	"fmt"
	"github.com/apache/arrow/go/arrow/memory"
)

const (
	INDEXOFFSET  = 0
	ENCLENOFFSET = 4
	DATAOFFSET   = 8
)

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

func (smp *ShmProvider) initEncoderDecoder() {
	// leave 4 bytes for the length of the message
	smp.buffer = *memory.NewBufferBytes(smp.ipcBuffer[DATAOFFSET:])
}

// Waits for messages or cancellation
// Writes a 1 to the index to indicate that the message has been read
func (smp *ShmProvider) Read() ([]byte, error) {

	// loop forever
	smp.bufmu.Lock()
	defer func() {
		smp.bufmu.Unlock()
	}()
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

// Send function
// Writes a 1 to the index to indicate that the message has been written
func (smp *ShmProvider) Write(data []byte) error {

	binary.LittleEndian.PutUint32(smp.ipcBuffer[INDEXOFFSET:], uint32(0))
	encodingLen := uint32(len(data))
	binary.LittleEndian.PutUint32(smp.ipcBuffer[ENCLENOFFSET:], uint32(encodingLen))
	copy(smp.ipcBuffer[DATAOFFSET:], data)

	// Signal the reader and wait for response
	smp.signalevent(smp.wrevent)

	return nil
}
