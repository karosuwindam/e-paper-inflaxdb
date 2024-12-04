TAG = 0.0.1
DOCKER = docker
NAME = senser-egent

TEMPLATE = ./Dockerfile_tmp
TARGET = ./Dockerfile
TARGET_FILE = ./
GO_VERSION = 1.22.7
ARCH = arm64
OPT = "--privileged"

TRACER_ON = true
TRACER_GRPC_URL = otel-grpc.bookserver.home:4317


BUILD = buildctl
BUILD_ADDR = tcp://buildkit.bookserver.home:1234 #arm64
BUILD_OPTION = "type=image,push=true,registry.insecure=true"

all: build
create:
	@echo "--- ${NAME} ${TAG} create ---"
	@echo "--- create Dockerfile ---"
	@cat ${TEMPLATE} | sed s/TAG/${TAG}/ | sed s/ARCH/${ARCH}/ | sed s/GO_VERSION/${GO_VERSION}/ > ${TARGET}
build: create
	@echo "--- build Dockerfile --"
	${DOCKER} build -t ${NAME}:${TAG} -f ${TARGET} .
build-kit: create
	@echo "--- buildkit build --"
	${BUILD} --addr ${BUILD_ADDR} build --output name=${NAME}:${TAG},${BUILD_OPTION} --frontend=dockerfile.v0 --local context=${TARGET_FILE}   --local dockerfile=${TARGET_FILE} --opt source=${TARGET}
rm: 
	${DOCKER} rmi ${NAME}:${TAG}
run:
	${DOCKER} run --rm --name=${NAME} ${OPT} -e TRACER_ON=${TRACER_ON} -e TRACER_GRPC_URL=${TRACER_GRPC_URL} ${NAME}:${TAG}
push:
	${DOCKER} push ${NAME}:${TAG}
