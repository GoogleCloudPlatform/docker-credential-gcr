# The file name of the binary to output
BINARY_FILENAME := docker-credential-gcr
# The output directory
OUT_DIR := bin
# The directory to dump generated mocks
MOCK_DIR := mock

all: clean bin

updatedeps: deps
	@go get -u github.com/kardianos/govendor
	@govendor add +external
	@govendor update +v

deps: bindeps testdeps

bindeps:
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
	@mkdir -p ${MOCK_DIR}/mock_config
	@mkdir -p ${MOCK_DIR}/mock_dockercmd
	@mockgen -destination=${MOCK_DIR}/mock_store/mocks.go github.com/GoogleCloudPlatform/docker-credential-gcr/store GCRCredStore
	@mockgen -destination=${MOCK_DIR}/mock_config/mocks.go github.com/GoogleCloudPlatform/docker-credential-gcr/config UserConfig
	@mockgen -destination=${MOCK_DIR}/mock_dockercmd/mocks.go github.com/GoogleCloudPlatform/docker-credential-gcr/util/dockercmd DockerClient

test: clean mocks bin
	@go test -timeout 10s -v -tags="unit integration surface" ./...

unit-tests: mocks
	@go test -timeout 10s -v -tags=unit ./...

integration-tests:
	@go test -timeout 10s -v -tags=integration ./...

surface-tests: bin
	@go test -timeout 10s -v -tags=surface ./...

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

presubmit: criticism pretty test
