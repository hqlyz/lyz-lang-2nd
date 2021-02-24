package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		oprands  []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.oprands...)
		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d", len(tt.expected), len(instruction))
		}

		for k, v := range tt.expected {
			if instruction[k] != tt.expected[k] {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d", k, v, instruction[k])
			}
		}
	}
}
