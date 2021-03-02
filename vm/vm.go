package vm

import (
	"fmt"
	"lyz-lang-2nd/code"
	"lyz-lang-2nd/compiler"
	"lyz-lang-2nd/object"
)

const (
	StackSize  = 2048
	GlobalSize = 65536
	MaxFrames  = 1024
)

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

// VM object
type VM struct {
	constants  []object.Object
	stack      []object.Object
	sp         int // Always points to the next value. Top of stack is stack[sp-1]
	globals    []object.Object
	frames     []*Frame
	frameIndex int
}

// New creates an instance of vm
func New(bytecode *compiler.Bytecode) *VM {
	mainFunc := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFunc, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:  bytecode.Constants,
		stack:      make([]object.Object, StackSize),
		sp:         0,
		globals:    make([]object.Object, GlobalSize),
		frames:     frames,
		frameIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

// Run method means power on the vm
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()

		op = code.Opcode(ins[ip])
		switch op {
		case code.OpConstant:
			constantIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constantIdx])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return nil
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			cond := vm.pop()
			if !isTruthy(cond) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpGetGlobal:
			index := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[index])
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			index := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			vm.globals[index] = vm.pop()
		case code.OpArray:
			num := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			arr := make([]object.Object, num)
			for i := num - 1; i >= 0; i-- {
				arr[i] = vm.pop()
			}
			err := vm.push(&object.Array{Elements: arr})
			if err != nil {
				return err
			}
		case code.OpHash:
			num := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			hash := make(map[object.HashKey]object.HashPair)
			for i := 0; i < num; i += 2 {
				v := vm.pop()
				k := vm.pop()
				kp := object.HashPair{Key: k, Value: v}
				hashKey, ok := k.(object.Hashable)
				if !ok {
					return fmt.Errorf("unusable as hash key: %s", k.Type())
				}
				hash[hashKey.HashKey()] = kp
			}
			err := vm.push(&object.Hash{Pairs: hash})
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			vm.currentFrame().ip++
			fn, ok := vm.stack[vm.sp-1].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function")
			}
			frame := NewFrame(fn, vm.sp)
			vm.pushFrame(frame)
			vm.sp = frame.basePointer + fn.NumLocals
		case code.OpReturnValue:
			returnValue := vm.pop()
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			frame := vm.currentFrame()
			vm.stack[frame.basePointer + int(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			err := vm.push(vm.stack[vm.currentFrame().basePointer + int(localIndex)])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(left, index object.Object) error {
	elems := left.(*object.Array).Elements
	i := index.(*object.Integer).Value

	max := int64(len(elems) - 1)
	if i < 0 || i > max {
		return vm.push(Null)
	}
	return vm.push(elems[i])
}

func (vm *VM) executeHashIndex(left, index object.Object) error {
	hashObject := left.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	return vm.push(pair.Value)
}

func (vm *VM) executeMinusOperator() error {
	value := vm.pop()
	if value.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("minus operator only support integer, got=%s", value.Type())
	}
	integer := value.(*object.Integer)
	return vm.push(&object.Integer{Value: -integer.Value})
}

func (vm *VM) executeBangOperator() error {
	value := vm.pop()
	switch value {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	} else if leftType == object.STRING_OBJ && rightType == object.STRING_OBJ {
		return vm.executeBinaryStringOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, leftObj object.Object, rightObj object.Object) error {
	var result int64
	leftValue := leftObj.(*object.Integer).Value
	rightValue := rightObj.(*object.Integer).Value
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, leftObj, rightObj object.Object) error {
	var result string
	leftValue := leftObj.(*object.String).Value
	rightValue := rightObj.(*object.String).Value
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return vm.push(&object.String{Value: result})
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	vm.sp--
	return vm.stack[vm.sp]
}

// LastPoppedStackElem is a test-only method
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func nativeBoolToBooleanObject(b bool) object.Object {
	if b {
		return True
	}
	return False
}

func isTruthy(cond object.Object) bool {
	switch cond {
	case True:
		return true
	case False:
		return false
	case Null:
		return false
	default:
		return true
	}
}
