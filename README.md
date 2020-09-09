# tgocc

Toy C compiler written in Go. Runs on Ubunto 18.04. Take a look at [Dockefile](https://github.com/joehattori/tgocc/blob/master/Dockerfile)!

# Usage
`tgocc` emits assembly for the given C file.
```
$ make
$ ./tgocc <file>.c > tmp.s
$ gcc -no-pie -o tmp tmp.s
$ ./tmp
```

# TODO
*`tgocc` is still under development. Any positive pull request is appreciated!*

One big goal is to be able to read gcc header files such as `stdio.h`, but this is only possible after implementing many other features of C and GNU extensions.
(Still, functions like `printf` and `strcmp` is available without `include`.)

Also the error messages are still poor and needs some improvement.

# Reference
Using [chibicc](https://github.com/rui314/chibicc) as reference.
