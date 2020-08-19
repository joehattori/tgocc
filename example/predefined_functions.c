int ret3() {
    return 3;
}

int id(int n) {
    return n;
}

int add(int a, int b) {
    return a + b;
}

int sumof6(int a, int b, int c, int d, int e, int f) {
    return a + b + c + d + e + f;
}

int *alloc4(int a, int b, int c, int d) {
    static int arr[4];
    arr[0] = a;
    arr[1] = b;
    arr[2] = c;
    arr[3] = d;
    return arr;
}
