GOBUILDARGS := -ldflags='-s -w'
BIN := ./bin
BINNAME = minidns

build:
	mkdir -p ${BIN}
	go build ${GOBUILDARGS} -o ${BIN}/${BINNAME} .

clean:
	rm -rf ${BIN}