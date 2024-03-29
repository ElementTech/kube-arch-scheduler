# Copyright 2019 The elementtech Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BIN_DIR=_output/bin

# If tag not explicitly set in users default to the git sha.
TAG ?= ${shell (git describe --tags --abbrev=14 | sed "s/-g\([0-9a-f]\{14\}\)$/+\1/") 2>/dev/null || git rev-parse --verify --short HEAD}

.EXPORT_ALL_VARIABLES:

all: local

init:
	mkdir -p ${BIN_DIR}

local: init
	go build -o=${BIN_DIR}/kube-arch-scheduler .

build-linux: init
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o=${BIN_DIR}/kube-arch-scheduler .

image: build-linux
	docker build --no-cache . -t ghcr.io/elementtech/kube-arch-scheduler:$(TAG)

push: 
	docker push ghcr.io/elementtech/kube-arch-scheduler:$(TAG)

update:
	go mod download
	go mod tidy
	go mod vendor

clean:
	rm -rf _output/
	rm -f *.log
