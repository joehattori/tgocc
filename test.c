// test written in C code.
/* Compile this code,
 * and see if it passes! */

int printf();
int exit();

int test(long expected, long actual, char *input) {
    if (actual == expected) {
        printf("%s => %ld\n", input, actual);
    } else {
        printf("%s => %ld expected, but got %ld\n", input, expected, actual);
        exit(1);
    }
    return 0;
}

int ret2() {
    return 2;
    return 1;
}

int add2(int x, int y) { return x+y; }
int sub2(int x, int y) { return x-y; }
int add6(int a, int b, int c, int d, int e, int f) { return a+b+c+d+e+f; }
int fib(int x) { if (x<=1) return 1; return fib(x-1) + fib(x-2); }
int addx(int *x, int y) { return *x+y; }
int sub_char(char a, char b, char c) { return a-b-c; }
int sub_short(short a, short b, short c) { return a-b-c; }
int sub_long(long a, long b, long c) { return a-b-c; }

int g1;
int g2[4];

int *gref() {
    return &g1;
}

char *ret_string() {
    return "test";
}

typedef long long ll;

static int static_fn() { return 3; }

int param_decay(int x[]) { return x[0]; }

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

    test(3, ({ int x=2; if (1) x=3; x; }), "int x=2; if (1) x=3; x;");
    test(3, ({ int x=0; if (2-2) x=2; else x=3; x; }), "int x=0; if (0) x=2; else x=3; x;");
    test(4, ({ int x=1; if (1==3) x=5; else if (1) x=4; x; }), "int x=1; if (1==3) x=5; else if (1) x=4; x;");
    test(3, ({ int x=({ int x=0; if (0) x=2; else x=3; x; }); x; }), "int x=({ int x=0; if (0) x=2; else x=3; x; }); x;");
    test(4, ({ int a=2; int x; if (a==3) x=5; else if (a==2) x=4; else x=3; x; }),
        "int a=2; int x; if (a==3) x=5; else if (a==2) x=4; else x=3; x;");
    test(3, ({ int x=5; if (0) x=2; else if (1-1) x=1; else x=3; x; }), "int x=5; if (0) x=2; else if (1-1) x=1; else x=3; x;");

    test(0, ({ int t = 100; while (t) t = t - 1; t; }), "int t = 100; while (t) t = t - 1; t;");
    test(10, ({ int t = 0; while (t != 10) t = t + 1; t; }), "int t = 0; while (t != 10) t = t + 1; return t;");

    test(10, ({ int t=100; for(t=0; t<10; t=t+1) 1; t; }), "int t=100; for(t=0; t<10; t=t+1) 1; t;");
    test(80, ({ int t=100; int i; for(i=0; i<10; i=i+1) {t=t-2;} t; }), "int t=100; int i; for(i=0; i<10; i=i+1) t=t-2; t;");
    test(3, ({ int a=0; int i; for(i=0; i<10; i=i+1) if(i==3) a=i; a; }),
        "int a=0; int i; for(i=0; i<10; i=i+1) if(i==3) a=i; a;");
    test(15, ({ int i=0; int v=0; for (; i<10; i=i+2) v=v+3; v; }),
        "int i=0; int v=0; for(; i<10; i=i+2) v=v+3; v;");
    test(55, ({ int i=0; int j=0; for (i=0; i<=10; i=i+1) j=i+j; j; }), "int i=0; int j=0; for (i=0; i<=10; i=i+1) j=i+j; j;");

    test(8, add2(3, 5), "add2(3, 5)");
    test(2, sub2(5, 3), "sub2(5, 3)");
    test(21, add6(1,2,3,4,5,6), "add6(1,2,3,4,5,6)");
    test(55, fib(9), "fib(9)");

    test(3, ({ int x=3; *&x; }), "int x=3; *&x;");
    test(3, ({ int x=3; int *y=&x; int **z=&y; **z; }), "int x=3; int *y=&x; int **z=&y; **z;");
    test(2, ({ int x=3; (&x+2)-&x; }), "int x=3; (&x+2)-&x;");

    test(5, ({ int x=3; int *y=&x; *y=5; x; }), "int x=3; int *y=&x; *y=5; x;");
    test(8, ({ int x=3; int y=5; addx(&x, y); }), "int x=3; int y=5; addx(&x, y);");

    test(3, ({ int x[2]; int *y=&x; *y=3; *x; }), "int x[2]; int *y=&x; *y=3; *x;");

    test(3, ({ int x[3]; *x=3; *(x+1)=4; *(x+2)=5; *x; }), "int x[3]; *x=3; *(x+1)=4; *(x+2)=5; *x;");
    test(4, ({ int x[3]; *x=3; *(x+1)=4; *(x+2)=5; *(x+1); }), "int x[3]; *x=3; *(x+1)=4; *(x+2)=5; *(x+1);");
    test(5, ({ int x[3]; *x=3; *(x+1)=4; *(x+2)=5; *(x+2); }), "int x[3]; *x=3; *(x+1)=4; *(x+2)=5; *(x+2);");

    test(0, ({ int x[2][3]; int *y=x; *y=0; **x; }), "int x[2][3]; int *y=x; *y=0; **x;");
    test(1, ({ int x[2][3]; int *y=x; *(y+1)=1; *(*x+1); }), "int x[2][3]; int *y=x; *(y+1)=1; *(*x+1);");
    test(2, ({ int x[2][3]; int *y=x; *(y+2)=2; *(*x+2); }), "int x[2][3]; int *y=x; *(y+2)=2; *(*x+2);");
    test(3, ({ int x[2][3]; int *y=x; *(y+3)=3; **(x+1); }), "int x[2][3]; int *y=x; *(y+3)=3; **(x+1);");
    test(4, ({ int x[2][3]; int *y=x; *(y+4)=4; *(*(x+1)+1); }), "int x[2][3]; int *y=x; *(y+4)=4; *(*(x+1)+1);");
    test(5, ({ int x[2][3]; int *y=x; *(y+5)=5; *(*(x+1)+2); }), "int x[2][3]; int *y=x; *(y+5)=5; *(*(x+1)+2);");
    test(6, ({ int x[2][3]; int *y=x; *(y+6)=6; **(x+2); }), "int x[2][3]; int *y=x; *(y+6)=6; **(x+2);");

    test(3, ({ int x[3]; *x=3; x[1]=4; x[2]=5; *x; }), "int x[3]; *x=3; x[1]=4; x[2]=5; *x;");
    test(4, ({ int x[3]; *x=3; x[1]=4; x[2]=5; *(x+1); }), "int x[3]; *x=3; x[1]=4; x[2]=5; *(x+1);");
    test(5, ({ int x[3]; *x=3; x[1]=4; x[2]=5; *(x+2); }), "int x[3]; *x=3; x[1]=4; x[2]=5; *(x+2);");
    test(5, ({ int x[3]; *x=3; x[1]=4; x[2]=5; *(x+2); }), "int x[3]; *x=3; x[1]=4; x[2]=5; *(x+2);");
    test(5, ({ int x[3]; *x=3; x[1]=4; 2[x]=5; *(x+2); }), "int x[3]; *x=3; x[1]=4; 2[x]=5; *(x+2);");

    test(0, ({ int x[2][3]; int *y=x; y[0]=0; x[0][0]; }), "int x[2][3]; int *y=x; y[0]=0; x[0][0];");
    test(1, ({ int x[2][3]; int *y=x; y[1]=1; x[0][1]; }), "int x[2][3]; int *y=x; y[1]=1; x[0][1];");
    test(2, ({ int x[2][3]; int *y=x; y[2]=2; x[0][2]; }), "int x[2][3]; int *y=x; y[2]=2; x[0][2];");
    test(3, ({ int x[2][3]; int *y=x; y[3]=3; x[1][0]; }), "int x[2][3]; int *y=x; y[3]=3; x[1][0];");
    test(4, ({ int x[2][3]; int *y=x; y[4]=4; x[1][1]; }), "int x[2][3]; int *y=x; y[4]=4; x[1][1];");
    test(5, ({ int x[2][3]; int *y=x; y[5]=5; x[1][2]; }), "int x[2][3]; int *y=x; y[5]=5; x[1][2];");
    test(6, ({ int x[2][3]; int *y=x; y[6]=6; x[2][0]; }), "int x[2][3]; int *y=x; y[6]=6; x[2][0];");

    test(4, ({ int x; sizeof(x); }), "int x; sizeof(x);");
    test(4, ({ int x; sizeof x; }), "int x; sizeof x;");
    test(8, ({ int *x; sizeof(x); }), "int *x; sizeof(x);");
    test(16, ({ int x[4]; sizeof(x); }), "int x[4]; sizeof(x);");
    test(48, ({ int x[3][4]; sizeof(x); }), "int x[3][4]; sizeof(x);");
    test(16, ({ int x[3][4]; sizeof(*x); }), "int x[3][4]; sizeof(*x);");
    test(4, ({ int x[3][4]; sizeof(**x); }), "int x[3][4]; sizeof(**x);");
    test(5, ({ int x[3][4]; sizeof(**x) + 1; }), "int x[3][4]; sizeof(**x) + 1;");
    test(5, ({ int x[3][4]; sizeof **x + 1; }), "int x[3][4]; sizeof **x + 1;");
    test(4, ({ int x[3][4]; sizeof(**x + 1); }), "int x[3][4]; sizeof(**x + 1);");

    test(0, g1, "g1");
    g1=3;
    test(3, g1, "g1");

    g2[0]=0; g2[1]=1; g2[2]=2; g2[3]=3;
    test(0, g2[0], "g2[0]");
    test(1, g2[1], "g2[1]");
    test(2, g2[2], "g2[2]");
    test(3, g2[3], "g2[3]");

    test(4, sizeof(g1), "sizeof(g1)");
    test(16, sizeof(g2), "sizeof(g2)");

    test(1, ({ char x=1; x; }), "char x=1; x;");
    test(1, ({ char x=1; char y=2; x; }), "char x=1; char y=2; x;");
    test(2, ({ char x=1; char y=2; y; }), "char x=1; char y=2; y;");

    test(1, ({ char x; sizeof(x); }), "char x; sizeof(x);");
    test(10, ({ char x[10]; sizeof(x); }), "char x[10]; sizeof(x);");
    test(1, sub_char(7, 3, 3), "sub_char(7, 3, 3)");

    test(97, "abc"[0], "\"abc\"[0]");
    test(98, "abc"[1], "\"abc\"[1]");
    test(99, "abc"[2], "\"abc\"[2]");
    test(0, "abc"[3], "\"abc\"[3]");
    test(3, sizeof("abc"), "sizeof(\"abc\")");

    test(2, ({ int x=2; { int x=3; } x; }), "int x=2; { int x=3; } x;");
    test(2, ({ int x=2; { int x=3; } int y=4; x; }), "int x=2; { int x=3; } int y=4; x;");
    test(3, ({ int x=2; { x=3; } x; }), "int x=2; { x=3; } x;");

    test(2, ({ struct { int first; int second; } s; s.first=2; s.first; }),
        "struct { int first; int second; } s; s.first=2; s.first");
    test(12, ({ struct {int a[3];} x; sizeof(x); }), "struct {int a[3];} x; sizeof(x);");
    test(16, ({ struct {int a;} x[4]; sizeof(x); }), "struct {int a;} x[4]; sizeof(x);");
    test(24, ({ struct {int a[3];} x[2]; sizeof(x); }), "struct {int a[3];} x[2]; sizeof(x)};");
    test(2, ({ struct {char a; char b;} x; sizeof(x); }), "struct {char a; char b;} x; sizeof(x);");
    test(8, ({ struct {char a; int b;} x; sizeof(x); }), "struct {char a; int b;} x; sizeof(x);");
    test(8, ({ struct {int a; char b;} x; sizeof(x); }), "struct {int a; char b;} x; sizeof(x);");

    test(1, ({ int x; char y; long a=&x; long b=&y; a-b; }), "int x; char y; long a=&x; long b=&y; a-b;");
    test(7, ({ char x; int y; long a=&x; long b=&y; a-b; }), "char x; int y; long a=&x; long b=&y; a-b;");

    test(8, ({ struct t {int a; int b;} x; struct t y; sizeof(y); }), "struct t {int a; int b;} x; struct t y; sizeof(y);");
    test(8, ({ struct t {int a; int b;}; struct t y; sizeof(y); }), "struct t {int a; int b;}; struct t y; sizeof(y);");
    test(2, ({ struct t {char a[2];}; { struct t {char a[4];}; } struct t y; sizeof(y); }),
        "struct t {char a[2];}; { struct t {char a[4];}; } struct t y; sizeof(y);");
    test(5, ({ struct t {int x;}; int t=2; struct t y; y.x=3; t+y.x; }),
        "struct t {int x;}; int t=2; struct t y; y.x=3; t+y.x;");

    test(3, ({ struct t {char a;} x; struct t *y=&x; x.a=3; y->a; }), "struct t {char a;} x; struct t *y=&x; x.a=3; y->a;");
    test(3, ({ struct t {char a;} x; struct t *y=&x; y->a=3; x.a; }), "struct t {char a;} x; struct t *y=&x; y->a=3; x.a;");

    test(2, ({ short x; sizeof(x); }), "short x; sizeof(x);");
    test(4, ({ struct {char a; short b;} x; sizeof(x); }), "struct {char a; short b;} x; sizeof(x);");

    test(8, ({ long x; sizeof(x); }), "long x; sizeof(x);");
    test(16, ({ struct {char a; long b;} x; sizeof(x); }), "struct {char a; long b} x; sizeof(x);");

    test(1, sub_short(7, 3, 3), "sub_short(7, 3, 3)");
    test(1, sub_long(7, 3, 3), "sub_long(7, 3, 3)");

    test(1, ({ typedef int t; t x=1; x; }), "typedef int t; t x=1; x;");
    test(1, ({ typedef struct {int a;} t; t x; x.a=1; x.a; }), "typedef struct {int a;} t; t x; x.a=1; x.a;");
    test(2, ({ typedef struct {int a;} t; { typedef int t; } t x; x.a=2; x.a; }), "typedef struct {int a;} t; { typedef int t; } t x; x.a=2; x.a;");

    test(3, *gref(), "*gref()");
    test(8, sizeof gref(), "sizeof gref();");
    test(116, ret_string()[0], "ret_string()[0];");
    test(115, ret_string()[2], "ret_string()[2];");

    test(24, ({ int *x[3]; sizeof(x); }), "int *x[3]; sizeof(x);");
    test(8, ({ int (*x)[3]; sizeof(x); }), "int (*x)[3]; sizeof(x);");
    test(3, ({ int *x[3]; int y; x[0]=&y; y=3; x[0][0]; }), "int *x[3]; int y; x[0]=&y; y=3; x[0][0];");
    test(4, ({ int x[3]; int (*y)[3]=x; y[0][0]=4; y[0][0]; }), "int x[3]; int (*y)[3]=x; y[0][0]=4; y[0][0];");

    test(1, sizeof(void), "sizeof(void)");
    test(1, sizeof(char), "sizeof(char)");
    test(2, sizeof(short), "sizeof(short)");
    test(2, sizeof(short int), "sizeof(short int)");
    test(4, sizeof(int), "sizeof(int)");
    test(8, sizeof(long), "sizeof(long)");
    test(8, sizeof(long int), "sizeof(long int)");
    test(8, sizeof(long long), "sizeof(long long)");
    test(8, sizeof(long long int), "sizeof(long long int)");

    test(8, sizeof(ll), "sizeof(ll)");
    test(3, ({ ll x=3; *&x; }), "ll x=3; *&x;");

    { void *x; }

    test(0, ({ _Bool x=0; x; }), "_Bool x=0; x;");
    test(1, ({ _Bool x=1; x; }), "_Bool x=1; x;");
    test(1, ({ _Bool x=2; x; }), "_Bool x=2; x;");

    test(131585, (int)8590066177, "(int)8590066177");
    test(513, (short)8590066177, "(short)8590066177");
    test(1, (char)8590066177, "(char)8590066177");
    test(1, (_Bool)1, "(_Bool)1");
    test(1, (_Bool)2, "(_Bool)2");
    test(0, (_Bool)(char)256, "(_Bool)(char)256");
    test(1, (long)1, "(long)1");
    test(0, (long)&*(int *)0, "(long)&*(int *)0");
    test(5, ({ int x=5; long y=(long)&x; *(int*)y; }), "int x=5; long y=(long)&x; *(int*)y");

    test(97, 'a', "'a'");
    test(122, 'z', "'z'");

    test(0, ({ enum { zero, one, two }; zero; }), "enum { zero, one, two }; zero;");
    test(1, ({ enum { zero, one, two, }; one; }), "enum { zero, one, two }; one;");
    test(2, ({ enum { zero, one, two }; two; }), "enum { zero, one, two }; two;");
    test(5, ({ enum { five=5, six, seven }; five; }), "enum { five=5, six, seven }; five;");
    test(6, ({ enum { five=5, six, seven, }; six; }), "enum { five=5, six, seven }; six;");
    test(0, ({ enum { zero, five=5, three=3, four, }; zero; }), "enum { zero, five=5, three=3, four }; zero;");
    test(5, ({ enum { zero, five=5, three=3, four, }; five; }), "enum { zero, five=5, three=3, four }; five;");
    test(3, ({ enum { zero, five=5, three=3, four }; three; }), "enum { zero, five=5, three=3, four }; three;");
    test(4, ({ enum { zero, five=5, three=3, four }; four; }), "enum { zero, five=5, three=3, four }; four;");
    test(4, ({ enum { zero, one, two } x; sizeof(x); }), "enum { zero, one, two } x; sizeof(x);");
    test(4, ({ enum t { zero, one, two, }; enum t y; sizeof(y); }), "enum t { zero, one, two }; enum t y; sizeof(y);");

    test(3, static_fn(), "static_fn()");

    test(55, ({ int j=0; for (int i=0; i<=10; i=i+1) j=j+i; j; }), "int j=0; for (int i=0; i<=10; i=i+1) j=j+i; j;");
    test(3, ({ int i=3; int j=0; for (int i=0; i<=10; i=i+1) j=j+i; i; }),
        "int i=3; int j=0; for (int i=0; i<=10; i=i+1) j=j+i; i;");

    test(3, ({ int i=2; ++i; }), "int i=2; ++i;");
    test(1, ({ int i=2; --i; }), "int i=2; --i;");
    test(2, ({ int i=2; i++; }), "int i=2; i++;");
    test(2, ({ int i=2; i--; }), "int i=2; i--;");
    test(3, ({ int i=2; i++; i; }), "int i=2; i++; i;");
    test(1, ({ int i=2; i--; i; }), "int i=2; i--; i;");
    test(1, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; *p++; }), "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; *p++;");
    test(2, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; ++*p; }), "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; ++*p;");
    test(1, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; *p--; }), "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; *p--;");
    test(0, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; --*p; }), "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; --*p;");

    test(0, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++)--; a[0]; }),
        "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++); a[0];");
    test(0, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++)--; a[1]; }),
        "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++); a[0];");
    test(2, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++)--; a[2]; }),
        "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++); a[0];");
    test(2, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++)--; *p; }),
        "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *p=a+1; (*p++); a[0];");

    test(7, ({ int i=2; i+=5; i; }), "int i=2; i+=5; i;");
    test(7, ({ int i=2; i+=5; }), "int i=2; i+=5;");
    test(3, ({ int i=5; i-=2; i; }), "int i=5; i-=2; i;");
    test(3, ({ int i=5; i-=2; }), "int i=5; i-=2;");
    test(6, ({ int i=3; i*=2; i; }), "int i=3; i*=2; i;");
    test(6, ({ int i=3; i*=2; }), "int i=3; i*=2;");
    test(3, ({ int i=6; i/=2; i; }), "int i=6; i/=2; i;");
    test(3, ({ int i=6; i/=2; }), "int i=6; i/=2;");
    test(2, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *b=a; b+=2; *b; }),
        "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *b=a; b+=2; *b;");
    test(1, ({ int a[3]; a[0]=0; a[1]=1; a[2]=2; int *b=a+2; b-=1; *b; }),
        "int a[3]; a[0]=0; a[1]=1; a[2]=2; int *b=a+2; b-=1; *b;");

    test(0, !1, "!1");
    test(0, !2, "!2");
    test(1, !0, "!0");

    test(3, ({ int i=0; for(;i<10;i++) { if (i==3) break; } i; }), "int i=0; for(;i<10;i++) { if (i==3) break; } i;");
    test(4, ({ int i=0; while (1) { if (i++ ==3) break; } i; }), "int i=0; while { if (i++ ==3) break; } i;");
    test(3, ({ int i=0; for(;i<10;i++) { for (;;) break; if (i==3) break; } i; }),
        "int i=0; for(;i<10;i++) { if (i==3) break; } i;");
    test(4, ({ int i=0; while (1) { while(1) break; if (i++ ==3) break; } i; }), "int i=0; while { if (i++ ==3) break; } i;");

    test(10, ({ int i=0; int j=0; for (;i<10;i++) { if (i>5) continue; j++; } i; }),
        "int i=0; int j=0; for (;i<10;i++) { if (i>5) continue; j++; } i;");
    test(6, ({ int i=0; int j=0; for (;i<10;i++) { if (i>5) continue; j++; } j; }),
        "int i=0; int j=0; for (;i<10;i++) { if (i>5) continue; j++; } j;");
    test(10, ({ int i=0; int j=0; for(;!i;) { for (;j!=10;j++) continue; break; } j; }),
        "int i=0; int j=0; for(;!i;) { for (;j!=10;j++) continue; break; } j;");
    test(11, ({ int i=0; int j=0; while (i++<10) { if (i>5) continue; j++; } i; }),
        "int i=0; int j=0; while (i++<10) { if (i>5) continue; j++; } i;");
    test(5, ({ int i=0; int j=0; while (i++<10) { if (i>5) continue; j++; } j; }),
        "int i=0; int j=0; while (i++<10) { if (i>5) continue; j++; } j;");
    test(11, ({ int i=0; int j=0; while(!i) { while (j++!=10) continue; break; } j; }),
        "int i=0; int j=0; while(!i) { while (j++!=10) continue; break; } j;");

    test(0, 0&1, "0&1");
    test(1, 3&1, "3&1");
    test(3, 7&3, "7&3");
    test(10, -1&10, " -1&10");

    test(511, 0777, "0777");
    test(0, 0x0, "0x0");
    test(10, 0xa, "0xa");
    test(10, 0XA, "0XA");
    test(48879, 0xbeef, "0xbeef");
    test(48879, 0xBEEF, "0xBEEF");
    test(48879, 0XBEEF, "0XBEEF");
    test(0, 0b0, "0b0");
    test(1, 0b1, "0b1");
    test(47, 0b101111, "0b101111");
    test(47, 0B101111, "0B101111");

    test(1, 0|1, "0|1");
    test(0b10011, 0b10000|0b00011, "01000|0b0011");

    test(0, 0^0, "0^0");
    test(0, 0b1111^0b1111, "0b1111^0b1111");
    test(0b110100, 0b111000^0b001100, "0b111000^0b001100");

    test(1, 0||1, "0||1");
    test(1, 0||(2-2)||5, "0||(2-2)||5");
    test(0, 0||0, "0||0");
    test(0, 0||(2-2), "0||(2-2)");

    test(0, 0&&1, "0&&1");
    test(0, (2-2)&&5, "(2-2)&&5");
    test(1, 1&&5, "1&&5");

    test(3, ({ int x[2]; x[0]=3; param_decay(x); }), "int x[2]; x[0]=3; param_decay(x);");

    printf("OK\n");
    return 0;
}
