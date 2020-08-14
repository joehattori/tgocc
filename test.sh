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

assert "0;" 0
assert "42;" 42

assert "1+3;" 4
assert "2-1;" 1
assert "5*3;" 15
assert "6/3;" 2

assert "2 + 5 * 3;" 17
assert "(2 + 5) * 3;" 21
assert "4 - (-1);" 5

assert "1==1;" 1
assert "1==3;" 0
assert "3!=5;" 1
assert "3!=3;" 0
assert "3<5;" 1
assert "3<3;" 0
assert "3<=3;" 1
assert "3<=2;" 0
assert "7>5;" 1
assert "2>3;" 0
assert "3>=3;" 1
assert "3>=9;" 0

assert "foo = 5; foo;" 5
assert "foo = 5; bar = 2; foo * bar;" 10
assert "f_oo2 = 1; bar = 2; f_oo2 = 4; f_oo2 + bar;" 6

assert "return 10;" 10
assert "foo = 5; bar = 2; return foo + bar;" 7
assert "return 10; return 20;" 10

assert "
if (1)
    3;" 3
assert "
if (1)
    return 5;" 5
assert "
if (0)
    return 5;
else
    return 4;" 4
assert "
if (1 == 3)
    return 5;
else if (1)
    return 4;" 4
assert "
if (3 + 2 == 6)
    return 5;
else if (1)
    return 4;
else return 3;" 4
assert "
a = 3;
if (a == 3)
    return 5;
else if (a == 2)
    return 4;
else return 3;" 5
assert "
if (0)
    return 5;
else if (1 == 2)
    return 4;
else return 3;" 3

assert "
t = 100;
while (t)
    t = t - 1;
return t;" 0
assert "
t = 0;
while (t != 10)
    t = t + 1;
return t;" 10

assert "
t = 100;
for (t = 0; t < 10; t = t+1)
    1;
return t;" 10
assert "
t = 100;
for (i = 0; i < 10; i = i+1)
    t = t - 2;
return t;" 80
assert "
a = 0;
for (i = 0; i < 10; i = i + 1)
    if (i == 3)
        a = i;
return a;" 3
assert "
for (i = 0;; i = i+1)
    if (i == 5)
        return i;
" 5
assert "
i = 0;
v = 0;
for (; i < 10; i = i+2)
    v = v + 3;
return v;
" 15
assert "
v = 0;
for (i = 0; i < 10;)
    if (i > 4)
        return i;
    else
        i = i + 2;
" 6

assert "
a = 0;
{ a = 42; }
return a;
" 42
assert "i=0; j=0; while(i<=10) {j=i+j; i=i+1;} return j;" 55
assert "
a = 0;
b = 0;
if (a == 0) {
    a = 10;
    b = 4;
}
return a+b;
" 14
assert "
s = 0;
a = 0;
for (i = 0; i < 10; i = i + 1) {
    s = s + i;
    a = i;
}
return a + s;
" 54
assert "
b = 0;
for(a=0; a <= 10; a=a+1) {
  if (a==6) {
    b=b+1;
    b=b+1;
    b=b+1;
  }
}
return b;
" 3
assert "
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
" 6

assert "
return return1();
" 1
assert "
return return3();
" 3

echo OK
