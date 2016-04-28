package ptrace

import "runtime"

var (
	ptraceChan     = make(chan func())
	ptraceDoneChan = make(chan struct{})
)

func init() {
	go handlePtraceRequests()
}

func OnPtraceThread(fn func()) {
	execOnPtraceThread(fn)
}

func handlePtraceRequests() {
	// We must ensure here that we are running on the same thread during
	// while invoking the ptrace(2) syscall. This is due to the fact that ptrace(2) expects
	// all commands after PTRACE_ATTACH to come from the same thread.
	runtime.LockOSThread()

	for fn := range ptraceChan {
		fn()
		ptraceDoneChan <- struct{}{}
	}
}

func execOnPtraceThread(fn func()) {
	ptraceChan <- fn
	<-ptraceDoneChan
}
