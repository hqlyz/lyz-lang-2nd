package main

import (
	"flag"
	"fmt"
	"lyz-lang-2nd/compiler"
	"lyz-lang-2nd/evaluator"
	"lyz-lang-2nd/lexer"
	"lyz-lang-2nd/object"
	"lyz-lang-2nd/parser"
	"lyz-lang-2nd/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")
var input = `
let fibonacci = fn(x) {
if (x == 0) {
0
} else {
if (x == 1) {
return 1;
} else {
fibonacci(x - 1) + fibonacci(x - 2);
}
}
};
fibonacci(35);
`

func main() {
	flag.Parse()
	var duration time.Duration
	var result object.Object
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}
		machine := vm.New(comp.Bytecode())
		start := time.Now()
		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}
		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		result = evaluator.Eval(program, env)
		duration = time.Since(start)
	}
	fmt.Printf(
		"engine=%s, result=%s, duration=%s\n",
		*engine,
		result.Inspect(),
		duration)

	s := time.Now()
	r := fibonacci(35)
	fmt.Printf(
		"engine=%s, result=%d, duration=%s\n",
		"go",
		r,
		time.Since(s))
}

func fibonacci(x int) int {
	if x <= 1 {
		return x
	}
	return fibonacci(x - 1) + fibonacci(x - 2)
}