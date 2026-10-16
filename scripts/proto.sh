#!/usr/bin/env bash
set -euo pipefail

# Generate the Go and TypeScript code for every .proto under proto/.
# Usage: proto.sh
#
# The toolchain is pinned along with the schema: protoc stamps its version into every
# generated Go and TypeScript file, so an unpinned generator turns a regeneration with
# no schema change into a diff. Each tool's version is asserted before anything runs.
#
# Go lands in backend/internal/proto (the go_package of each file, under the backend
# module); TypeScript in frontend/src/generated. ts-proto generates generic definitions
# rather than grpc-js services, matching @primandproper/platform-client, whose
# transports carry them; grpc-js definitions call Buffer at runtime and cannot run in a
# browser. Both output trees are replaced whole, so a deleted message leaves nothing
# behind.

PROTOC_VERSION="33.1"
PROTOC_GEN_GO_VERSION="1.36.4"
PROTOC_GEN_GO_GRPC_VERSION="1.5.1"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_MODULE="github.com/primandproper/nosmallparts/backend"
GO_OUT="${ROOT}/backend/internal/proto"
TS_OUT="${ROOT}/frontend/src/generated"
TS_PLUGIN="${ROOT}/node_modules/.bin/protoc-gen-ts_proto"

assert_version() {
  local tool="$1" want="$2" fix="$3" got
  if ! got="$("${tool}" --version 2>&1)"; then
    echo "proto.sh: ${tool} is not installed; ${fix}" >&2
    exit 1
  fi
  if [ "${got}" != "${want}" ]; then
    echo "proto.sh: want ${want}, have ${got}; ${fix}" >&2
    exit 1
  fi
}

assert_version protoc "libprotoc ${PROTOC_VERSION}" \
  "install protoc ${PROTOC_VERSION} (mise install protoc@${PROTOC_VERSION}) or put it first on PATH"
assert_version protoc-gen-go "protoc-gen-go v${PROTOC_GEN_GO_VERSION}" \
  "go install google.golang.org/protobuf/cmd/protoc-gen-go@v${PROTOC_GEN_GO_VERSION}"
assert_version protoc-gen-go-grpc "protoc-gen-go-grpc ${PROTOC_GEN_GO_GRPC_VERSION}" \
  "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v${PROTOC_GEN_GO_GRPC_VERSION}"
if [ ! -x "${TS_PLUGIN}" ]; then
  echo "proto.sh: ${TS_PLUGIN} is missing; run pnpm install" >&2
  exit 1
fi

protos=()
while IFS= read -r -d '' proto; do
  protos+=("${proto#"${ROOT}/proto/"}")
done < <(find "${ROOT}/proto" -name '*.proto' -print0 | sort -z)

if [ ${#protos[@]} -eq 0 ]; then
  echo "proto.sh: no .proto files under ${ROOT}/proto" >&2
  exit 1
fi

work="$(mktemp -d)"
# shellcheck disable=SC2064
trap "rm -rf '${work}'" EXIT
mkdir -p "${work}/go" "${work}/ts"

(
  cd "${ROOT}/proto"
  protoc \
    --go_out="${work}/go" --go_opt=module="${BACKEND_MODULE}" \
    --go-grpc_out="${work}/go" --go-grpc_opt=module="${BACKEND_MODULE}" \
    --plugin=protoc-gen-ts_proto="${TS_PLUGIN}" \
    --ts_proto_out="${work}/ts" \
    --ts_proto_opt=outputServices=generic-definitions \
    --ts_proto_opt=esModuleInterop=true \
    --proto_path . \
    "${protos[@]}"
)

rm -rf "${GO_OUT}" "${TS_OUT}"
mkdir -p "$(dirname "${GO_OUT}")" "$(dirname "${TS_OUT}")"
mv "${work}/go/internal/proto" "${GO_OUT}"
mv "${work}/ts" "${TS_OUT}"

echo "generated $(find "${GO_OUT}" -name '*.go' | wc -l | tr -d ' ') Go and $(find "${TS_OUT}" -name '*.ts' | wc -l | tr -d ' ') TypeScript files from ${#protos[@]} protos"
