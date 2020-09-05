#define ZERO 0
#define WEEKS 365/7
#define MIN(X, Y)  ((X) < (Y) ? (X) : (Y))

int printf();
int exit();
int strcmp(char *p, char *q);
int add2(int x, int y);
int sub2(int x, int y);
int add6(int a, int b, int c, int d, int e, int f);
int fib(int x);
int addx(int *x, int y);
int sub_char(char a, char b, char c);
int sub_short(short a, short b, short c);
int sub_long(long a, long b, long c);

typedef struct {
    int a;
    int b;
    int c;
} abc;
