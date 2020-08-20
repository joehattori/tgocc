#!/bin/bash

cc -xc -c -o tmp2.o example/predefined_functions.c

assert() {
    input="$1"
    expected="$2"

    ./tgocc "$input" > tmp.s
    cc -static -o tmp tmp.s tmp2.o
    ./tmp
    actual="$?"

    if [ "$actual" = "$expected" ]; then
        echo "$input => $actual"
    else
        echo "$input => $expected expected, but got $actual"
        exit 1
    fi
}

assert "int main() { return 0; }" 0
assert "int main() { return 42; }" 42

assert "int main() { return 1+3; }" 4
assert "int main() { return 2-1; }" 1
assert "int main() { return 5*3; }" 15
assert "int main() { return 6/3; }" 2

assert "int main() { return 2 + 5 * 3; }" 17
assert "int main() { return (2 + 5) * 3; }" 21
assert "int main() { return 4 - (-1); }" 5

assert "int main() { return 1==1; }" 1
assert "int main() { return 1==3; }" 0
assert "int main() { return 3!=5; }" 1
assert "int main() { return 3!=3; }" 0
assert "int main() { return 3<5; }" 1
assert "int main() { return 3<3; }" 0
assert "int main() { return 3<=3; }" 1
assert "int main() { return 3<=2; }" 0
assert "int main() { return 7>5; }" 1
assert "int main() { return 2>3; }" 0
assert "int main() { return 3>=3; }" 1
assert "int main() { return 3>=9; }" 0

assert "int main() { int foo; foo = 5; return foo; }" 5
assert "int main() { int foo = 5; int bar = 2; return foo * bar; }" 10
assert "int main() { int f_oo2 = 1; int bar = 2; f_oo2 = 4; return f_oo2 + bar; }" 6
assert "int main() { return 10; return 20; }" 10
assert "int main() { 10; return 20; }" 20

assert "int main() { if (1) return 3; }" 3
assert "int main() { if (0) return 5; else return 4; }" 4
assert "int main() { if (1 == 3) return 5; else if (1) return 4; }" 4
assert "int main() { if (3 + 2 == 6) return 5; else if (1) return 4; else return 3; }" 4
assert "int main() { int a = 3; if (a == 3) return 5; else if (a == 2) return 4; else return 3; }" 5
assert "int main() { if (0) return 5; else if (1 == 2) return 4; else return 3; }" 3

assert "int main() { int t = 100; while (t) t = t - 1; return t; }" 0
assert "int main() { int t = 0; while (t != 10) t = t + 1; return t; }" 10

assert "int main() { int t = 100; for (t = 0; t < 10; t = t+1) 1; return t; }" 10
assert "int main() { int t = 100; int i; for (i = 0; i < 10; i = i+1) t = t - 2; return t; }" 80
assert "int main() { int a = 0; int i; for (i = 0; i < 10; i = i + 1) if (i == 3) a = i; return a; }" 3
assert "int main() { int i; for (i = 0;; i = i+1) if (i == 5) return i; }" 5
assert "int main() { int i = 0; int v = 0; for (; i < 10; i = i+2) v = v + 3; return v; }" 15
assert "int main() { int v = 0; int i; for (i = 0; i < 10;) if (i > 4) return i; else i = i + 2; }" 6

assert "int main() { int a = 0; { a = 42; } return a; }" 42
assert "int main() { int i = 0; int j = 0; while(i<=10) {j=i+j; i=i+1;} return j; }" 55
assert "int main() { int a = 0; int b = 0; if (a == 0) { a = 10; b = 4; } return a+b; }" 14
assert "int main() { int s = 0; int a = 0; int i; for (i = 0; i < 10; i = i + 1) { s = s + i; a = i; } return a + s; }" 54
assert "int main() { int b = 0; int a; for(a=0; a <= 10; a=a+1) { if (a==6) { b=b+3; } } return b; }" 3

assert "int main() { return ret3(); }" 3
assert "int main() { return id(2); }" 2
assert "int main() { return add(100, 111); }" 211
assert "int main() { return sumof6(1, 2, 3, 4, 5, 6); }" 21

assert "int ret5() { return 5; } int main() { return ret5(); } " 5
assert "int idn(int n) { return n; } int main() { return idn(1); } " 1
assert "
int sum6(int a, int b, int c, int d, int e, int f) { return a+b+c+d+e+f; }
int main() { return sum6(1,2,3,4,5,6); } " 21
assert " int fib(int n) { if (n <= 1) return n; return fib(n - 1) + fib(n - 2); } int main() { return fib(10); } " 55

assert "int main() { int *x; *x = 4; return *x; } " 4
assert "int main() { int *x; int y = 4; *x = y; return *x; } " 4
assert "int main() { int x = 3; int *y = &x; return *y; }" 3
assert "int main() { int x = 3; int *y = &x; x = 4; return *y; }" 4
assert "int main() { int x; int *y = &x; *y = 3; return x; }" 3
assert "int *retref() { int *y; *y = 33; return y; } int main() { int *x = retref(); return *x; }" 33
assert "int main() { int *y; *y = 4; int **x; x = &y; return **x; } " 4
assert "int main() { int x=3; return *&x; }" 3
assert "int main() { int x=3; int *y=&x; int **z; z=&y; return **z; }" 3

assert "int main() { int *p = alloc4(1,2,100,4); return *p; }" 1
assert "int main() { int *p = alloc4(1,2,100,4); int *q; q = p + 2; return *q; }" 100

assert "int main() { int x = 34; return sizeof(x); }" 4
assert "int main() { int *x; *x = 20; return sizeof(x); }" 8
assert "int main() { int x = 3; return sizeof(x + 1); }" 4
assert "int main() { int *x; *x = 20; return sizeof(*x); }" 4
assert "int main() { return sizeof(1); }" 4
assert "int main() { int *x; *x = 3; return sizeof(sizeof(*x)); }" 4

assert "int main() { int i[10]; *i = 1; *(i+1) = 2; return *i; }" 1
assert "int main() { int i[10]; *i = 1; *(i+1) = 2; return *(i+1); }" 2
assert "int main() { int i[10]; *i = 1; *(i+1) = 2; int *p = i; return *(p + 1); }" 2
assert "int main() { int i[10]; return sizeof(i); }" 40

assert "int main() { int i[10]; i[0] = 100; return *i; }" 100
assert "int main() { int i[10]; i[0] = 1; i[3] = 2; return i[0]+i[3]; }" 3
assert "int main() { int *i=alloc4(2,3,4,5); return i[0]; }" 2
assert "int main() { int *i=alloc4(2,3,4,5); int s=0; int k; for(k=0;k<4;k=k+1) s=s+i[k]; return s; }" 14

echo OK
