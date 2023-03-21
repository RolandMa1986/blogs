

## python go  交互方式

1. api 调用 grpc,webapi,etc
2. pipe 管道
3. 直接调用？

开发成本
调用效率
兼容性

## cpython

1. 扩展一个 c 模块

目标：

```python
import spam
status = spam.system("ls -l")
```

必须包含的头文件
```c
#define PY_SSIZE_T_CLEAN
#include <Python.h>
```

实现方法
```c
static PyObject *
spam_system(PyObject *self, PyObject *args)
{
    const char *command;
    int sts;

    if (!PyArg_ParseTuple(args, "s", &command))
        return NULL;
    sts = system(command);
    return PyLong_FromLong(sts);
}
```
注册方法
``` 
static struct PyModuleDef spammodule = {
    PyModuleDef_HEAD_INIT,
    "spam",   /* name of module */
    spam_doc, /* module documentation, may be NULL */
    -1,       /* size of per-interpreter state of the module,
                 or -1 if the module keeps state in global variables. */
    SpamMethods
};

static PyMethodDef SpamMethods[] = {
    ...
    {"system",  spam_system, METH_VARARGS,
     "Execute a shell command."},
    ...
    {NULL, NULL, 0, NULL}        /* Sentinel */
};

PyMODINIT_FUNC
PyInit_spam(void)
{
    return PyModule_Create(&spammodule);
}

```

2. 集成 python interrapter


### 阅读

https://docs.python.org/3/extending/extending.html#a-simple-example

## cgo

什么是cgo

### go 调用 c 库

1. 内嵌代码
```go
package main
/*
#include <stdio.h>
int SayHello() {
 puts("Hello World");
    return 0;
}
*/
import "C"
import (
    "fmt"
)

func main() {
    ret := C.SayHello()
    fmt.Println(ret)
}
```


### c 调用 go

```go
// demo/hello.go
package main

//#include <hello.h>
import "C"
import "fmt"

// ### export 为关键字
//export SayHello
func SayHello(str *C.char) {
    fmt.Println(C.GoString(str))
}
```

### 技巧
1. 查看编译命令

2. pkg-config 依赖包管理

3.  

### 阅读

https://zhuanlan.zhihu.com/p/349197066

## go-python

1. go 调用 python 方法

2. python 调用 go 方法



## 处理

1. 线程安全

2. 类型转换

3. 异常处理

4. 托管内存

5. Windows/linux 跨平台

6. python 版本特性/兼容性

