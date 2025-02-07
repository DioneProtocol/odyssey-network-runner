#!/usr/bin/env bash
set -e

export RUN_E2E="true"

# e.g.,
# ./scripts/tests.e2e.sh $VERSION1 $SUBNET_EVM_VERSION
if ! [[ "$0" =~ scripts/tests.e2e.sh ]]; then
    echo "must be run from repository root"
    exit 255
fi

ONR_PATH=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

DEFAULT_VERSION_1=v1.10.10
DEFAULT_SUBNET_EVM_VERSION=v0.5.6

if [ $# == 0 ]; then
    VERSION_1=$DEFAULT_VERSION_1
    SUBNET_EVM_VERSION=$DEFAULT_SUBNET_EVM_VERSION
else
    VERSION_1=$1
    if [[ -z "${VERSION_1}" ]]; then
        echo "Missing version argument!"
        echo "Usage: ${0} [VERSION_1] [SUBNET_EVM_VERSION]" >>/dev/stderr
        exit 255
    fi
    SUBNET_EVM_VERSION=$3
    if [[ -z "${SUBNET_EVM_VERSION}" ]]; then
        echo "Missing version argument!"
        echo "Usage: ${0} [VERSION_1] [SUBNET_EVM_VERSION]" >>/dev/stderr
        exit 255
    fi
fi

echo "Running e2e tests with:"
echo VERSION_1: ${VERSION_1}
echo SUBNET_EVM_VERSION: ${SUBNET_EVM_VERSION}

#
# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"

############################
ODYSSEYGO_REPO=/tmp/odysseygo/
if [ ! -d $ODYSSEYGO_REPO ]; then
    git clone https://github.com/DioneProtocol/odysseygo $ODYSSEYGO_REPO
fi

CORETH_REPO=/tmp/coreth/
if [ ! -d $CORETH_REPO ]; then
    git clone https://github.com/DioneProtocol/coreth $CORETH_REPO
fi

VERSION_1_DIR=/tmp/odysseygo-${VERSION_1}/
if [ ! -f ${VERSION_1_DIR}/ ]; then
    echo building $VERSION_1
    rm -rf ${VERSION_1_DIR}
    mkdir -p ${VERSION_1_DIR}
    cd $ODYSSEYGO_REPO
    git checkout $VERSION_1
    ./scripts/build.sh
    cp -r build/* ${VERSION_1_DIR}
fi

SUBNET_EVM_REPO=/tmp/subnet-evm-repo/
if [ ! -d $SUBNET_EVM_REPO ]; then
    git clone https://github.com/DioneProtocol/subnet-evm $SUBNET_EVM_REPO
fi

SUBNET_EVM_VERSION_DIR=/tmp/subnet-evm-${SUBNET_EVM_VERSION}/
if [ ! -f $VERSION_1_DIR/plugins/srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy ]; then
    echo building subnet-evm $SUBNET_EVM_VERSION
    rm -rf ${SUBNET_EVM_VERSION_DIR}
    mkdir -p ${SUBNET_EVM_VERSION_DIR}
    cd $SUBNET_EVM_REPO
    git checkout $SUBNET_EVM_VERSION
    # NOTE: We are copying the subnet-evm binary here to a plugin hardcoded as srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy which corresponds to the VM name `subnetevm` used as such in the test
    ./scripts/build.sh $VERSION_1_DIR/plugins/srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy
fi

############################

cd $ONR_PATH

echo "building runner"
./scripts/build.sh

echo "building e2e.test"
# to install the ginkgo binary (required for test build and run)
go install -v github.com/onsi/ginkgo/v2/ginkgo@v2.1.3
ACK_GINKGO_RC=true ginkgo build ./tests/e2e
./tests/e2e/e2e.test --help

snapshots_dir=/tmp/network-runner-root-data/snapshots-e2e/
rm -rf $snapshots_dir

killall odyssey-network-runner || true

echo "launch local test cluster in the background"
bin/odyssey-network-runner \
    server \
    --log-level debug \
    --port=":8080" \
    --snapshots-dir=$snapshots_dir \
    --grpc-gateway-port=":8081" &
#--disable-nodes-output \
PID=${!}

function cleanup() {
    echo "shutting down network runner"
    kill ${PID}
}
trap cleanup EXIT

echo "running e2e tests"
./tests/e2e/e2e.test \
    --ginkgo.v \
    --ginkgo.fail-fast \
    --log-level debug \
    --grpc-endpoint="0.0.0.0:8080" \
    --grpc-gateway-endpoint="0.0.0.0:8081" \
    --odysseygo-path=${VERSION_1_DIR}/odysseygo
