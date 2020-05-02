.PHONY: codegen
codegen:
	./codegen/update.sh

.PHONY: verify-codegen
verify-codegen:
	./codegen/verify.sh

.PHONY: dependency
dependency:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod verify
	@GO111MODULE=on go mod vendor

.PHONY: spot-manager
spot-manager:
	@rm bin/spot-manager || true
	@GO111MODULE=on CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/spot-manager ./app/spot-manager

.PHONY: spot-cluster
spot-cluster:
	@rm bin/spot-cluster || true
	@GO111MODULE=on CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/spot-cluster ./app/spot-cluster

.PHONY: spot-manager-image
spot-manager-image: spot-manager
	@docker build -t shovan1995/spot-manager:latest -f package/Dockerfile .

.PHONY: binary
binary: spot-manager spot-cluster
