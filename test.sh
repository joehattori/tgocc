#!/bin/bash

assert() {
    input="$1"
    expected="$2"

    ./tgocc "$input" > tmp.s
    cc -o tmp tmp.s
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

assert "if (1) 3;" 3
assert "if (1) return 5;" 5
assert "if (0) return 5; else return 4;" 4
assert "if (1 == 3) return 5; else if (1) return 4;" 4
assert "if (3 + 2 == 6) return 5; else if (1) return 4; else return 3;" 4
assert "if (0) return 5; else if (1 == 2) return 4; else return 3;" 3

echo OK
