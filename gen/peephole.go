package gen

import "github.com/evanphx/m13/insn"

type PeepholeOptz struct{}

func (p *PeepholeOptz) Optimize(stream []insn.Instruction) {
	for i, in := range stream {
		if in.Op() == insn.CopyReg {
			ik := stream[i+1]

			if ik.Op() == insn.CopyReg {
				if in.R0() == ik.R1() && in.R1() == ik.R0() {
					stream[i+1] = insn.Builder.Noop()
				}
			}
		}
	}
}
