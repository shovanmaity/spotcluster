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

.PHONY: spotmanager
spotmanager:
	@rm bin/spot-manager || true
	@GO111MODULE=on CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/spot-manager ./app/spot-manager

.PHONY: spotmanager-image
spotmanager-image: spotmanager
	@docker build -t shovan1995/spot-manager:ci -f package/Dockerfile .
