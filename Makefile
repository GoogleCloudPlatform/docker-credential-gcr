# The file name of the binary to output
BINARY_FILENAME := docker-credential-gcr
# The output directory
OUT_DIR := bin
# The directory to dump generated mocks
MOCK_DIR := mock

all: clean deps bin

test: clean testdeps unit-tests integration-tests

deps:
	@go get -u -t ./...
	
testdeps:
	@go get -u -t ./...
	@go get -u github.com/golang/mock/gomock
	@go get -u github.com/golang/mock/mockgen

bin:
	@go build -i -o ${OUT_DIR}/${BINARY_FILENAME} main.go
	@echo Binary created: ${OUT_DIR}/${BINARY_FILENAME}

clean:
	@rm -rf ${OUT_DIR}
	@rm -rf ${MOCK_DIR}
	@go clean

mocks:
	@rm -rf ${MOCK_DIR}
	@mkdir -p ${MOCK_DIR}/mock_store
	@mockgen -destination=${MOCK_DIR}/mock_store/mocks.go github.com/GoogleCloudPlatform/docker-credential-gcr/store GCRCredStore

unit-tests: mocks
	@go test -timeout 10s -v -tags=unit ./...
	
integration-tests:
	@go test -timeout 10s -v -tags=integration ./...
	
vet: 
	@go vet ./...

lint:
	@golint ./...
	
criticism: clean vet lint

fmt:
	@gofmt -w -s .
	
fix:
	@go fix ./...
	
pretty: fmt fix
