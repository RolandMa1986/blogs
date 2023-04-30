package main

import (
	"fmt"
	"log"
	"os"

	python3 "github.com/go-python/cpy3"
)

func main() {
	python3.Py_Initialize()
	defer python3.Py_Finalize()
	if !python3.Py_IsInitialized() {
		fmt.Println("Error initializing the python interpreter")
		os.Exit(1)
	}

	ret := python3.PyRun_SimpleString("import sys\nsys.path.append('/home/roland/roland/blogs/go-python/tutorial2')\n")
	if ret != 0 {
		log.Fatalf("error appending '%s' to python sys.path", "pyoutliers")
	}

	oImport := python3.PyImport_ImportModule("fibo")
	if !(oImport != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		log.Fatal("failed to add module 'fibo'")
	}
	defer oImport.DecRef()

	fib := oImport.GetAttrString("fib")
	if !(fib != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		log.Fatal("failed to add module 'fibo'")
	}
	defer fib.DecRef()

	sub_call(fib)

}

func sub_call(fib *python3.PyObject) {
	n := python3.PyLong_FromGoInt(5)
	args := python3.PyTuple_New(1)             //retval: New reference
	ret := python3.PyTuple_SetItem(args, 0, n) //steals ref to pylist
	if ret != 0 {
		if python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
		}
		fmt.Errorf("error setting args tuple item")
		return
	}
	testdataPy := fib.CallObject(args) //retval: New reference
	defer testdataPy.DecRef()
	n.DecRef()
	args.DecRef()
}
