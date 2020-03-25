GOBUILDARGS := -ldflags='-s -w'
BIN := ./bin
BINNAME = minidns

build:
	mkdir -p ${BIN}
	go build ${GOBUILDARGS} -o ${BIN}/${BINNAME} .

clean:
	rm -rf ${BIN}

run:
	MINIDNS_PORT=53 MINIDNS_BIND=192.168.1.174 MINIDNS_LOG_LEVEL=vverbose go run . 

docker:
	docker build -t tcfw/minidns:$(shell git rev-parse --short HEAD) .

docker-push:
	docker push tcfw/minidns:$(shell git rev-parse --short HEAD)