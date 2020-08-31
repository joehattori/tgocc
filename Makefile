SRCS=*.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./tgocc test.c > tmp.s
	echo 'int ext1; int *ext2; int char_fn() { return 257; }' | cc -xc -c -o tmp2.o -
	cc -static -o tmp tmp.s tmp2.o
	./tmp

test_old: tgocc
	cc -xc -c -o tmp2.o example/predefined_functions.c
	./test.sh

.PHONY: clean test tgocc
