package repl

import (
	"bufio"
	"fmt"
	"io"
	"lyz-lang-2nd/compiler"
	"lyz-lang-2nd/lexer"
	"lyz-lang-2nd/object"
	"lyz-lang-2nd/parser"
	"lyz-lang-2nd/vm"
)

// PROMPT is a console prompt symbol
const PROMPT = ">> "

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

// Start function
func Start(in io.Reader, out io.Writer) {
	scan := bufio.NewScanner(in)
	w := bufio.NewWriter(out)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	for {
		w.WriteString(PROMPT)
		w.Flush()
		scanned := scan.Scan()
		if !scanned {
			return
		}
		line := scan.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errs()) != 0 {
			printParserErrors(w, p.Errs())
			w.Flush()
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}
		// constants = code.Constants

		machine := vm.NewWithGlobalsStore(comp.Bytecode(), globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}
		stackTop := machine.LastPoppedStackElem()
		if stackTop != nil {
			w.WriteString(stackTop.Inspect())
			w.WriteString("\n")
		}
		w.Flush()
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
