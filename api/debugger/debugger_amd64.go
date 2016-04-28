package debugger

import (
	"debug/gosym"

	"github.com/derekparker/delve/api/debugger/disasm"

	"rsc.io/x86/x86asm"
)

type instrseq []x86asm.Op

var windowsPrologue = instrseq{x86asm.MOV, x86asm.MOV, x86asm.LEA, x86asm.CMP, x86asm.JBE}
var windowsPrologue2 = instrseq{x86asm.MOV, x86asm.MOV, x86asm.CMP, x86asm.JBE}
var unixPrologue = instrseq{x86asm.MOV, x86asm.LEA, x86asm.CMP, x86asm.JBE}
var unixPrologue2 = instrseq{x86asm.MOV, x86asm.CMP, x86asm.JBE}
var prologues = []instrseq{windowsPrologue, windowsPrologue2, unixPrologue, unixPrologue2}

// FirstPCAfterPrologue returns the address of the first instruction after the prologue for function fn
// If sameline is set FirstPCAfterPrologue will always return an address associated with the same line as fn.Entry
func (d *Debugger) FirstPCAfterPrologue(fn *gosym.Func, sameline bool) (uint64, error) {
	regs, err := d.ct.Registers()
	if err != nil {
		return 0, err
	}
	text, err := disasm.Between(fn.Entry, fn.End, d.ct.Mem, d.bp, regs, false)
	if err != nil {
		return fn.Entry, err
	}

	if len(text) <= 0 {
		return fn.Entry, nil
	}

	for _, prologue := range prologues {
		if len(prologue) >= len(text) {
			continue
		}
		if checkPrologue(text, prologue) {
			panic("figure out asm -> location")
			// r := &text[len(prologue)]
			// if sameline {
			// 	if r.Loc.Line != text[0].Loc.Line {
			// 		return fn.Entry, nil
			// 	}
			// }
			// return r.Loc.PC, nil
		}
	}

	return fn.Entry, nil
}

func checkPrologue(s []disasm.AsmInstruction, prologuePattern instrseq) bool {
	for i, op := range prologuePattern {
		if s[i].Inst.Op != op {
			return false
		}
	}
	return true
}
