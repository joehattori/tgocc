#!/bin/bash

cat <<EOF | cc -xc -c -o tmp2.o example/predefined_functions.c
EOF

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

assert "main() { return 0; }" 0
assert "main() { return 42; }" 42

assert "main() { return 1+3; }" 4
assert "main() { return 2-1; }" 1
assert "main() { return 5*3; }" 15
assert "main() { return 6/3; }" 2

assert "main() { return 2 + 5 * 3; }" 17
assert "main() { return (2 + 5) * 3; }" 21
assert "main() { return 4 - (-1); }" 5

assert "main() { return 1==1; }" 1
assert "main() { return 1==3; }" 0
assert "main() { return 3!=5; }" 1
assert "main() { return 3!=3; }" 0
assert "main() { return 3<5; }" 1
assert "main() { return 3<3; }" 0
assert "main() { return 3<=3; }" 1
assert "main() { return 3<=2; }" 0
assert "main() { return 7>5; }" 1
assert "main() { return 2>3; }" 0
assert "main() { return 3>=3; }" 1
assert "main() { return 3>=9; }" 0

assert "main() { foo = 5; return foo; }" 5
assert "main() { foo = 5; bar = 2; return foo * bar; }" 10
assert "main() { f_oo2 = 1; bar = 2; f_oo2 = 4; return f_oo2 + bar; }" 6
assert "main() { return 10; return 20; }" 10
assert "main() { 10; return 20; }" 20


assert "
main() {
    if (1)
        return 3;
}" 3

assert "
main() {
if (0)
    return 5;
else
    return 4;
}" 4

assert "
main() {
    if (1 == 3)
        return 5;
    else if (1)
        return 4;
}" 4

assert "
main() {
    if (3 + 2 == 6)
        return 5;
    else if (1)
        return 4;
    else return 3;
}" 4

assert "
main() {
    a = 3;
    if (a == 3)
        return 5;
    else if (a == 2)
        return 4;
    else return 3;
}" 5

assert "
main() {
    if (0)
        return 5;
    else if (1 == 2)
        return 4;
    else return 3;
}" 3


assert "
main() {
    t = 100;
    while (t)
        t = t - 1;
    return t;
}" 0

assert "
main() {
    t = 0;
    while (t != 10)
        t = t + 1;
    return t;
}" 10


assert "
main() {
    t = 100;
    for (t = 0; t < 10; t = t+1)
        1;
    return t;
}" 10

assert "
main() {
    t = 100;
    for (i = 0; i < 10; i = i+1)
        t = t - 2;
    return t;
}" 80

assert "
main() {
    a = 0;
    for (i = 0; i < 10; i = i + 1)
        if (i == 3)
            a = i;
    return a;
}" 3

assert "
main() {
    for (i = 0;; i = i+1)
        if (i == 5)
            return i;
}" 5

assert "
main() {
    i = 0;
    v = 0;
    for (; i < 10; i = i+2)
        v = v + 3;
    return v;
}" 15

assert "
main() {
    v = 0;
    for (i = 0; i < 10;)
        if (i > 4)
            return i;
        else
            i = i + 2;
}" 6


assert "
main() {
    a = 0;
    { a = 42; }
    return a;
}" 42

assert "
main() {
    i=0; j=0; while(i<=10) {j=i+j; i=i+1;} return j;
}" 55

assert "
main() {
    a = 0;
    b = 0;
    if (a == 0) {
        a = 10;
        b = 4;
    }
    return a+b;
}" 14

assert "
main() {
    s = 0;
    a = 0;
    for (i = 0; i < 10; i = i + 1) {
        s = s + i;
        a = i;
    }
    return a + s;
}" 54

assert "
main() {
    b = 0;
    for(a=0; a <= 10; a=a+1) {
      if (a==6) {
        b=b+1;
        b=b+1;
        b=b+1;
      }
    }
    return b;
}" 3

assert "
main() {
    b = 0;
    for (a=0; a <= 10; a=a+1) {
      if (a==2) {
        b=b+1;
        b=b+1;
      } else if (a == 4) {
        b=b+1;
        b=b+1;
        b=b+1;
        b=b+1;
      }
    }
    return b;
}" 6


assert "
main() {
    return ret3();
}" 3

assert "
main() {
    return id(2);
}" 2

assert "
main() {
    return add(100, 111);
}" 211

assert "
main() {
    return sumof6(1, 2, 3, 4, 5, 6);
}" 21


assert "
ret5() { return 5; }
main() { return ret5(); }
" 5

assert "
idn(n) { return n; }
main() { return idn(1); }
" 1

assert "
sum6(a,b,c,d,e,f) { return a+b+c+d+e+f; }
main() { return sum6(1,2,3,4,5,6); }
" 21

assert "
fib(n) {
    if (n <= 1)
        return n;
    return fib(n - 1) + fib(n - 2);
}
main() { return fib(10); }
" 55


assert "
main() {
    x = 3;
    y = &x;
    return *y;
}" 3

assert "
main() {
    x = 3;
    y = &x;
    x = 4;
    return *y;
}" 4

echo OK
