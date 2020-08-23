#!/bin/bash

assert() {
    input="$1"
    expected="$2"

    ./tgocc <(echo "$input") > tmp.s
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

echo OK
