package proc

import (
	"fmt"

	"github.com/derekparker/delve/proc/ptrace"

	sys "golang.org/x/sys/unix"
)

// OSSpecificDetails hold Linux specific
// process details.
type OSSpecificDetails struct {
	registers sys.PtraceRegs
}

func threadHalt(t *Thread) error {
	err := sys.Tgkill(t.p.Pid(), t.ID, sys.SIGSTOP)
	if err != nil {
		return fmt.Errorf("halt err %s on thread %d", err, t.ID)
	}
	// TODO(derekparker) check we got the right thread back...
	_, _, err = t.p.Wait()
	if err != nil {
		return fmt.Errorf("wait err %s on thread %d", err, t.ID)
	}
	return nil
}

func threadStopped(t *Thread) bool {
	state := status(t.ID, t.p.os.comm)
	return state == StatusTraceStop || state == StatusTraceStopT
}

func threadResume(t *Thread) error {
	return resumeWithSig(t, 0)
}

func resumeWithSig(t *Thread, sig int) error {
	t.running = true
	return ptrace.PtraceCont(t.ID, sig)
}

func (t *Thread) singleStep() error {
	for {
		err := ptrace.PtraceSingleStep(t.ID)
		if err != nil {
			return err
		}
		// TODO(derekparker) verify correct thread
		th, status, err := t.p.Wait()
		if err != nil {
			return err
		}
		if (status == nil || status.Exited()) && th.ID == t.p.Pid() {
			// t.p.postExit()
			rs := 0
			if status != nil {
				rs = status.ExitStatus()
			}
			return ProcessExitedError{Pid: t.p.Pid(), Status: rs}
		}
		if th.ID == t.ID && status.Signal() == sys.SIGTRAP {
			return nil
		}
	}
}

func (t *Thread) blocked() bool {
	pc, _ := t.PC()
	fn := t.p.goSymTable.PCToFunc(pc)
	if fn != nil && ((fn.Name == "runtime.futex") || (fn.Name == "runtime.usleep") || (fn.Name == "runtime.clone")) {
		return true
	}
	return false
}
