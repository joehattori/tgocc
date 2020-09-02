#include "test.h"

int add2(int x, int y) { return x+y; }
int sub2(int x, int y) { return x-y; }
int add6(int a, int b, int c, int d, int e, int f) { return a+b+c+d+e+f; }
int fib(int x) { if (x<=1) return 1; return fib(x-1) + fib(x-2); }
int addx(int *x, int y) { return *x+y; }
int sub_char(char a, char b, char c) { return a-b-c; }
int sub_short(short a, short b, short c) { return a-b-c; }
int sub_long(long a, long b, long c) { return a-b-c; }
