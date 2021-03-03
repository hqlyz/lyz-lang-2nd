package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

// Opcode is the operator in the bytecode used in vm
type Opcode byte

const (
	OpConstant       Opcode = iota // 0
	OpAdd                          // 1
	OpPop                          // 2
	OpSub                          // 3
	OpMul                          // 4
	OpDiv                          // 5
	OpTrue                         // 6
	OpFalse                        // 7
	OpEqual                        // 8
	OpNotEqual                     // 9
	OpGreaterThan                  // 10
	OpMinus                        // 11
	OpBang                         // 12
	OpJumpNotTruthy                // 13
	OpJump                         // 14
	OpNull                         // 15
	OpSetGlobal                    // 16
	OpGetGlobal                    // 17
	OpArray                        // 18
	OpHash                         // 19
	OpIndex                        // 20
	OpCall                         // 21
	OpReturnValue                  // 22
	OpReturn                       // 23
	OpSetLocal                     // 24
	OpGetLocal                     // 25
	OpGetBuiltin                   // 26
	OpClosure                      // 27
	OpGetFree                      // 28
	OpCurrentClosure               // 29
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant", []int{2}},
	OpAdd:            {"OpAdd", []int{}},
	OpPop:            {"OpPop", []int{}},
	OpSub:            {"OpSub", []int{}},
	OpMul:            {"OpMul", []int{}},
	OpDiv:            {"OpDiv", []int{}},
	OpTrue:           {"OpTrue", []int{}},
	OpFalse:          {"OpFalse", []int{}},
	OpEqual:          {"OpEqual", []int{}},
	OpNotEqual:       {"OpNotEqual", []int{}},
	OpGreaterThan:    {"OpGreaterThan", []int{}},
	OpMinus:          {"OpMinus", []int{}},
	OpBang:           {"OpBang", []int{}},
	OpJumpNotTruthy:  {"OpJumpNotTruthy", []int{2}},
	OpJump:           {"OpJump", []int{2}},
	OpNull:           {"OpNull", []int{}},
	OpSetGlobal:      {"OpSetGlobal", []int{2}},
	OpGetGlobal:      {"OpGetGlobal", []int{2}},
	OpArray:          {"OpArray", []int{2}},
	OpHash:           {"OpHash", []int{2}},
	OpIndex:          {"OpIndex", []int{}},
	OpCall:           {"OpCall", []int{1}},
	OpReturnValue:    {"OpReturnValue", []int{}},
	OpReturn:         {"OpReturn", []int{}},
	OpSetLocal:       {"OpSetLocal", []int{1}},
	OpGetLocal:       {"OpGetLocal", []int{1}},
	OpGetBuiltin:     {"OpGetBuiltin", []int{1}},
	OpClosure:        {"OpClosure", []int{2, 1}},
	OpGetFree:        {"OpGetFree", []int{1}},
	OpCurrentClosure: {"OpCurrentClosure", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)
	offset := 1
	for i, o := range operands {
		w := def.OperandWidths[i]
		switch w {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += w
	}

	return instruction
}

func ReadOperands(def *Definition, operands []byte) ([]int, int) {
	r := []int{}
	offset := 0
	for _, w := range def.OperandWidths {
		switch w {
		case 2:
			v := ReadUint16(operands[offset:])
			r = append(r, int(v))
		case 1:
			v := ReadUint8(operands[offset:])
			r = append(r, int(v))
		}
		offset += w
	}

	return r, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n\t", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}
	switch operandCount {
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 0:
		return def.Name
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
