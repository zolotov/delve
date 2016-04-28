package goroutine

import (
	"debug/dwarf"
	"encoding/binary"
	"fmt"
	"reflect"
	"runtime"

	"github.com/derekparker/delve/api/debugger/location"
	"github.com/derekparker/delve/proc"
)

// G status, from: src/runtime/runtime2.go
const (
	Gidle           uint64 = iota // 0
	Grunnable                     // 1 runnable and on a run queue
	Grunning                      // 2
	Gsyscall                      // 3
	Gwaiting                      // 4
	GmoribundUnused               // 5 currently unused, but hardcoded in gdb scripts
	Gdead                         // 6
	Genqueue                      // 7 Only the Gscanenqueue is used.
	Gcopystack                    // 8 in this state when newstack is moving the stack
)

// G represents a runtime G (goroutine) structure (at least the
// fields that Delve is interested in).
type G struct {
	ID         int    // Goroutine ID
	PC         uint64 // PC of goroutine when it was parked.
	SP         uint64 // SP of goroutine when it was parked.
	GoPC       uint64 // PC of 'go' statement that created this goroutine.
	WaitReason string // Reason for goroutine being parked.
	Status     uint64

	// Information on goroutine location
	CurrentLoc location.Location

	// PC of entry to top-most deferred function.
	DeferPC uint64

	// Thread that this goroutine is currently allocated to
	thread *proc.Thread

	dbp *proc.Process
}

func RunningOnThread(t *proc.Thread, dwarf *dwarf.Dwarf) (*G, error) {
	gaddr, deref, err := goroutineAddress(t)
	if err != nil {
		return nil, err
	}

	v, err := newGVariable(gaddr, deref, dwarf)
	if err != nil {
		return nil, err
	}

	// 	gaddr, err := thread.getGVariable()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	g, err = gaddr.parseG()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	g.thread = thread
	// 	return g, nil
}

// goroutineAddress will return the address of the G struct
// in the thread local storage for the given thread. Additionally
// it will return whether that address must be dereferenced when
// evaluating the G struct.
func goroutineAddress(t *proc.Thread) (uint64, bool, error) {
	goroutineTLSOffset := t.GStructOffset()
	if goroutineTLSOffset == 0 {
		// GetG was called through SwitchThread / updateThreadList during initialization
		// thread.dbp.arch isn't setup yet (it needs a CurrentThread to read global variables from)
		return nil, false, fmt.Errorf("g struct offset not initialized")
	}
	regs, err := t.Registers()
	if err != nil {
		return nil, false, err
	}
	gaddrbytes, err := t.Mem.Read(uintptr(regs.TLS()+goroutineTLSOffset, t.PtrSize()))
	if err != nil {
		return nil, false, err
	}
	gaddr := uintptr(binary.LittleEndian.Uint64(gaddrbytes))
	// On Windows, the value at TLS()+GStructOffset() is a
	// pointer to the G struct.
	deref := runtime.GOOS == "windows"
	return gaddr, deref, nil
}

func newGVariable(addr uint64, deref bool, dwarf *dwarf.Dwarf) (*Variable, error) {
	typ, err := dwarf.TypeNamed("runtime.g")
	if err != nil {
		return nil, err
	}

	name := ""

	if deref {
		typ = &dwarf.PtrType{dwarf.CommonType{int64(thread.dbp.arch.PtrSize()), "", reflect.Ptr, 0}, typ}
	} else {
		name = "runtime.curg"
	}

	return thread.newVariable(name, gaddr, typ), nil
}
