int assert(int expected, int actual, char *input) {
    if (actual == expected) {
        printf("%s => %d\n", input, actual);
    } else {
        printf("%s => %d expected, but got %d\n", input, expected, actual);
        exit(1);
    }
}

int ret2() {
    return 2;
    return 1;
}

int main() {
    assert(0, 0, "0");
    assert(42, 42, "42");

    assert(4, 1+3, "1+3");
    assert(1, 2-1, "2-1");
    assert(15, 5*3, "5*3");
    assert(2, 6/3, "6/3");
    assert(17, 2+5*3, "2+5*3");
    assert(21, (2+5)*3, "(2+5)*3");
    assert(5, 4-(-1), "4-(-1)");

    assert(1, 1==1, "1==1");
    assert(0, 1==3, "1==3");
    assert(1, 3!=5, "3!=5");
    assert(0, 3!=3, "3!=3");
    assert(1, 3<5, "3<5");
    assert(0, 3<3, "3<3");
    assert(1, 3<=3, "3<=3");
    assert(0, 5<=3, "5<=3");
    assert(1, 7>5, "7>5");
    assert(0, 2>3, "2>3");
    assert(1, 3>=3, "3>=3");
    assert(0, 3>=9, "3>=9");

    assert(8, ({ int x; int z; x=3; z=5; x+z; }), "int a; int z; a=3; z=5; a+z;");
    assert(3, ({ int x; x=3; x; }), "int x; x=3; x;");
    /*assert(12, ({ int x=3; int y=4; x+y; }), "int x=3; int y=4; x+y;");*/
    assert(6, ({ int f_oo2 = 1; int bar = 2; f_oo2 = 4; f_oo2 + bar; }), 6);

    printf("OK\n");
    return 0;
}
