#!/bin/bash

assert() {
    expected="$1"
    input="$2"

    ./tgocc "$input" > tmp.s
    cc -o tmp tmp.s
    ./tmp
    actual="$?"

    if [ "$actual"="$expected" ]; then
        echo "$input => $actual"
    else
        echo "$input => $expected expected, but got $actual"
        exit 1
    fi
}

assert 0 0
assert 42 42

assert "1+3" 4
assert "2-1" 1
assert "5*3" 15
assert "6/3" 2

assert "2 + 5 * 3" 17
assert "(2 + 5) * 3" 21
assert "4 - (-1)" 5

echo OK
