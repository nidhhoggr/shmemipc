//go:build !windows

package shmemipc

import (
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/nidhhoggr/go-arrow/arrow/memory"
)

type ShmProvider struct {
	name      string
	closed    bool
	bufmu     sync.Mutex
	ipcBuffer []byte
	buffer    memory.Buffer
	event     uintptr
	rdevent   uintptr
	wrevent   uintptr
}

const (
	IPC_CREAT int = 01000
	IPC_RMID  int = 0
	SETVAL    int = 16
)

type semop struct {
	semNum  uint16
	semOp   int16
	semFlag int16
}

func errnoErr(errno syscall.Errno) error {
	switch errno {
	case syscall.Errno(0):
		return nil
	default:
		return error(errno)
	}
}

func Ftok(path string, id uint) (uint, error) {
	st := &syscall.Stat_t{}
	if err := syscall.Stat(path, st); err != nil {
		return 0, err
	}
	return uint((uint(st.Ino) & 0xffff) | uint((st.Dev&0xff)<<16) |
		((id & 0xff) << 24)), nil
}

func (smp *ShmProvider) initevents() error {
	_, _, err := syscall.Syscall6(
		uintptr(syscall.SYS_SEMCTL),
		uintptr(smp.event),
		uintptr(0),
		uintptr(SETVAL),
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if err != syscall.Errno(0) {

		return err
	}
	_, _, err = syscall.Syscall6(
		uintptr(syscall.SYS_SEMCTL),
		uintptr(smp.event),
		uintptr(1),
		uintptr(SETVAL),
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if err != syscall.Errno(0) {

		return err
	}
	return nil
}

func (smp *ShmProvider) openevents(filename string, flag int) error {
	key, err := Ftok(filename, 0)
	if err != nil {
		return err
	}
	r1, _, err := syscall.Syscall(
		syscall.SYS_SEMGET,
		uintptr(key),
		uintptr(2),
		uintptr(flag))
	if err != syscall.Errno(0) {
		return err
	}
	smp.event = r1
	smp.rdevent = 0
	smp.wrevent = 1
	return nil
}

func (smp *ShmProvider) signalevent(event uintptr) error {
	post := semop{semNum: uint16(event), semOp: 1, semFlag: 0}
	_, _, err := syscall.Syscall(syscall.SYS_SEMOP, uintptr(smp.event),
		uintptr(unsafe.Pointer(&post)), uintptr(1))
	if err != syscall.Errno(0) {

		return errnoErr(err)
	}
	return nil
}

func (smp *ShmProvider) waitforevent(event uintptr) error {
	wait := semop{semNum: uint16(event), semOp: -1, semFlag: 0}
	_, _, err := syscall.Syscall(syscall.SYS_SEMOP, uintptr(smp.event),
		uintptr(unsafe.Pointer(&wait)), uintptr(1))
	if err != syscall.Errno(0) {

		return errnoErr(err)
	}
	return nil
}

func (smp *ShmProvider) Listen(filename string, len uint64, flags int) error {
	if flags == 0 {
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}
	f, err := os.OpenFile(filename, flags, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Truncate(int64(len))
	ptr, err := syscall.Mmap(int(f.Fd()), 0, int(len), syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return err
	}

	err = smp.openevents(filename, IPC_CREAT|0666)
	if err != nil {
		smp.Close(nil)
		return err
	}

	smp.name = filename
	smp.ipcBuffer = ptr
	smp.initEncoderDecoder()
	return nil
}

func (smp *ShmProvider) Dial(filename string) error {
	f, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	ptr, err := syscall.Mmap(int(f.Fd()), 0, int(stat.Size()), syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {

		return err
	}

	err = smp.openevents(filename, 0)
	if err != nil {

		return err
	}

	smp.initevents()
	smp.ipcBuffer = ptr
	smp.initEncoderDecoder()
	return nil
}

func (smp *ShmProvider) Close(wg *sync.WaitGroup) error {

	// signal waiting listening goroutine if there is one
	if wg != nil {
		wg.Wait()
	}

	smp.closed = true
	//why would it write here?
	//because it causes us to break out of the read loop
	smp.signalevent(smp.wrevent)

	smp.bufmu.Lock()
	defer smp.bufmu.Unlock()

	if smp.ipcBuffer != nil {

		syscall.Munmap(smp.ipcBuffer)
	}
	if smp.name != "" {

		// this is the server if we created the file and recorded its name
		_, _, _ = syscall.Syscall(syscall.SYS_SEMCTL, uintptr(smp.event),
			uintptr(0), uintptr(IPC_RMID))
		syscall.Unlink(smp.name)
	}
	return nil
}
