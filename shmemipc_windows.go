package shmemipc

import (
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	memory "github.com/nidhhoggr/go-arrow/arrow/memory"
	"golang.org/x/sys/windows"
)

const (
	eventRequestReadySuffix  = "-RequestReadyEvent"
	eventResponseReadySuffix = "-ResponseReadyEvent"
)

type ShmProvider struct {
	name      string
	closed    bool
	bufmu     sync.Mutex
	ipcBuffer []byte
	buffer    memory.Buffer
	handle    uintptr
	wrevent   uintptr
	rdevent   uintptr
}

// Create signalling event
func (smp *ShmProvider) createevents(name string) error {

	r1, err := windows.CreateEvent(nil, 0, /*auto-reset*/
		0 /*not signaled*/, UTF16PtrFromString(name+eventRequestReadySuffix))
	if err != nil {
		return err
	}
	smp.rdevent = uintptr(r1)
	r1, err = windows.CreateEvent(nil, 0, /*auto-reset*/
		0 /*not signaled*/, UTF16PtrFromString(name+eventResponseReadySuffix))
	if err != nil {
		return err
	}
	smp.wrevent = uintptr(r1)
	return nil
}

func (smp *ShmProvider) openevents(name string) error {

	r1, err := windows.OpenEvent(windows.SYNCHRONIZE|windows.EVENT_MODIFY_STATE,
		false, UTF16PtrFromString(name+eventRequestReadySuffix))
	if err != nil {
		return err
	}
	smp.rdevent = uintptr(r1)
	r1, err = windows.OpenEvent(windows.SYNCHRONIZE|windows.EVENT_MODIFY_STATE,
		false, UTF16PtrFromString(name+eventResponseReadySuffix))
	if err != nil {
		return err
	}
	smp.wrevent = uintptr(r1)
	return nil
}

func (smp *ShmProvider) waitforevent(event uintptr) {
	syscall.WaitForSingleObject(syscall.Handle(event), syscall.INFINITE)
}

func (smp *ShmProvider) signalevent(event uintptr) {
	windows.SetEvent(windows.Handle(event))
}

func (smp *ShmProvider) closeevents() {
	if smp.rdevent != 0 {
		syscall.CloseHandle(syscall.Handle(smp.rdevent))
	}
	if smp.wrevent != 0 {
		syscall.CloseHandle(syscall.Handle(smp.wrevent))
	}
}

func UTF16PtrFromString(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

// This exists because OpenFileMapping is not in the windows package (yet)
var procOpenFileMappingW = windows.NewLazySystemDLL("kernel32.dll").NewProc("OpenFileMappingW")

func OpenFileMapping(desiredAccess uint32, inheritHandle bool, name *uint16) (handle windows.Handle, err error) {
	var _p0 uint32
	if inheritHandle {
		_p0 = 1
	}
	r0, _, e1 := syscall.Syscall(procOpenFileMappingW.Addr(), 3, uintptr(desiredAccess), uintptr(_p0), uintptr(unsafe.Pointer(name)))
	handle = windows.Handle(r0)
	if handle == 0 {
		err = e1
	}
	return
}

// Listen Creates a file mapping with the specified name and size, and returns a handle to the file mapping.
func (smp *ShmProvider) Listen(name string, len uint64, flags int) (err error) {

	flagArgs := uint32(flags)
	if flagArgs == 0 {
		flagArgs = windows.FILE_MAP_READ | windows.FILE_MAP_WRITE
	}

	defer func() {
		if err != nil {
			smp.close() // If returning any error, make sure all native resources are destroyed
		}
	}()

	// Create the file mapping
	r1, err := windows.CreateFileMapping(windows.InvalidHandle, nil, windows.PAGE_READWRITE, uint32(len>>32), uint32(len&0xffffffff),
		UTF16PtrFromString(name))
	if err != nil {
		return err
	}
	smp.handle = uintptr(r1)

	// Map the file into memory
	ptr, err := windows.MapViewOfFile(windows.Handle(smp.handle), flagArgs, 0, 0, 0)
	if err != nil {
		return err
	}
	smp.ipcBuffer = unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len)

	// Create the event
	err = smp.createevents(name)
	if err != nil {
		return err
	}
	smp.initEncoderDecoder()
	smp.name = name
	runtime.SetFinalizer(smp, func(smp *ShmProvider) { smp.close() })
	return nil
}

// Dial Opens a file mapping with the specified name, and returns a handle to the file mapping.
func (smp *ShmProvider) Dial(name string) (err error) {
	defer func() {
		if err != nil {
			smp.close() // If returning any error, make sure all native resources are destroyed
		}
	}()
	r1, err := OpenFileMapping(windows.FILE_MAP_WRITE, false, UTF16PtrFromString(name))
	if err != nil {
		return err
	}
	smp.handle = uintptr(r1)
	ptr, err := windows.MapViewOfFile(windows.Handle(smp.handle), windows.FILE_MAP_READ|windows.FILE_MAP_WRITE, 0, 0, 0)
	if err != nil {
		return err
	}
	var mbi windows.MemoryBasicInformation
	err = windows.VirtualQuery(ptr, &mbi, unsafe.Sizeof(mbi))
	if err != nil {
		return err
	}
	smp.ipcBuffer = unsafe.Slice((*byte)(unsafe.Pointer(ptr)), mbi.RegionSize)

	// Create the events
	err = smp.openevents(name)
	if err != nil {
		return err
	}
	smp.initEncoderDecoder()
	return nil
}

func (smp *ShmProvider) close() { // Finalize
	smp.bufmu.Lock()
	defer smp.bufmu.Unlock()
	if smp.ipcBuffer != nil {
		windows.UnmapViewOfFile(uintptr(unsafe.Pointer(&smp.ipcBuffer[0])))
	}
	if smp.handle != 0 {
		syscall.CloseHandle(syscall.Handle(smp.handle))
	}
	smp.closeevents()
}

func (smp *ShmProvider) Close(wg *sync.WaitGroup) error {

	// signal waiting listening goroutine if there is one
	if wg != nil {
		smp.closed = true
		smp.signalevent(smp.wrevent)
		wg.Wait()
	}
	smp.close()
	return nil
}
