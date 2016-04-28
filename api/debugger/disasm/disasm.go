package disasm

import (
	"github.com/derekparker/delve/api/debugger/breakpoint"
	"github.com/derekparker/delve/proc"
	"github.com/derekparker/delve/proc/memory"
)

type AsmInstruction struct {
	PC         uint64
	CallPC     uint64
	Bytes      []byte
	Breakpoint bool
	AtPC       bool
	Inst       *ArchInst
}

type AssemblyFlavour int

const (
	GNUFlavour = AssemblyFlavour(iota)
	IntelFlavour
)

// Between disassembles target memory between startPC and endPC
// If currentGoroutine is set and thread is stopped at a CALL instruction Disassemble will evaluate the argument of the CALL instruction using the thread's registers
// Be aware that the Bytes field of each returned instruction is a slice of a larger array of size endPC - startPC
func Between(startPC, endPC uint64, mem memory.ReadWriter, bps breakpoint.Breakpoints, regs proc.Registers, currentGoroutine bool) ([]AsmInstruction, error) {
	rawbytes, err := mem.Read(startPC, int(endPC-startPC))
	if err != nil {
		return nil, err
	}

	r := make([]AsmInstruction, 0, len(rawbytes)/15)
	pc := startPC

	var curpc uint64
	if regs != nil {
		curpc = regs.PC()
	}

	for len(rawbytes) > 0 {
		bp, atbp := bps[pc]
		if atbp {
			for i := range bp.OriginalData {
				rawbytes[i] = bp.OriginalData[i]
			}
		}
		inst, err := asmDecode(rawbytes, pc)
		if err == nil {
			atpc := currentGoroutine && (curpc == pc)
			callpc := resolveCallArgPC(inst, atpc, mem, regs)
			r = append(r, AsmInstruction{
				PC:         pc,
				CallPC:     callpc,
				Bytes:      rawbytes[:inst.Len],
				Breakpoint: atbp,
				AtPC:       atpc,
				Inst:       inst,
			})

			pc += uint64(inst.Size())
			rawbytes = rawbytes[inst.Size():]
		} else {
			r = append(r, AsmInstruction{PC: pc, Bytes: rawbytes[:1], Breakpoint: atbp, Inst: nil})
			pc++
			rawbytes = rawbytes[1:]
		}
	}
	return r, nil
}
