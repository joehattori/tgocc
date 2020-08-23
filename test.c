int test(int expected, int actual, char *input) {
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
    test(0, 0, "0");
    test(42, 42, "42");

    test(4, 1+3, "1+3");
    test(1, 2-1, "2-1");
    test(15, 5*3, "5*3");
    test(2, 6/3, "6/3");
    test(17, 2+5*3, "2+5*3");
    test(21, (2+5)*3, "(2+5)*3");
    test(5, 4-(-1), "4-(-1)");

    test(1, 1==1, "1==1");
    test(0, 1==3, "1==3");
    test(1, 3!=5, "3!=5");
    test(0, 3!=3, "3!=3");
    test(1, 3<5, "3<5");
    test(0, 3<3, "3<3");
    test(1, 3<=3, "3<=3");
    test(0, 5<=3, "5<=3");
    test(1, 7>5, "7>5");
    test(0, 2>3, "2>3");
    test(1, 3>=3, "3>=3");
    test(0, 3>=9, "3>=9");

    test(3, ({ int x; x=3; x; }), "int x; x=3; x;");
    test(8, ({ int a; int b; a=3; b=5; a+b; }), "int a; int b; a=3; b=5; a+b;");
    test(6, ({ int f_oo2=1; int bar=2; f_oo2=4; f_oo2+bar; }), "int f_oo2=1; int bar=2; f_oo2=4; f_oo2+bar;");

    printf("OK\n");
    return 0;
}
