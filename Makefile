SRCS=*.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./tgocc test/test1.c > tmp1.s
	cc -xc -c test/util.c -o tmp_util.o
	cc -xc -c test/test2.c -o tmp2.o
	cc -static -o tmp tmp1.s tmp_util.o tmp2.o
	./tmp

.PHONY: clean test tgocc
