package memory

import "github.com/derekparker/delve/proc/ptrace"

func read(tid int, addr uint64, size int) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}
	return ptrace.PtracePeekData(tid, uintptr(addr), size)
}

func write(tid int, addr uint64, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	return ptrace.PtracePokeData(tid, uintptr(addr), data)
}
