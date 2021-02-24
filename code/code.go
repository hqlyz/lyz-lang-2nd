package code

import (
	"encoding/binary"
	"fmt"
)

type Instructions []byte

// Opcode is the operator in the bytecode used in vm
type Opcode byte

const (
	OpConstant Opcode = iota
)

type Definition struct {
	Name         string
	OprandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, oprands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OprandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)
	offset := 1
	for i, o := range oprands {
		w := def.OprandWidths[i]
		switch w {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += w
	}

	return instruction
}
