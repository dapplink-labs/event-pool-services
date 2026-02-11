RRM_ABI_ARTIFACT := ./abis/ReferralRewardManager.sol/ReferralRewardManager.json

event-pool-services:
	env GO111MODULE=on go build -v $(LDFLAGS) ./cmd/event-pool-services

clean:
	rm event-pod-services

test:
	go test -v ./...

lint:
	golangci-lint run ./...

bindings: binding-rrm

binding-rrm:
	$(eval temp := $(shell mktemp))

	cat $(RRM_ABI_ARTIFACT) \
		| jq -r .bytecode.object > $(temp)

	cat $(RRM_ABI_ARTIFACT) \
		| jq .abi \
		| abigen --pkg bindings \
		--abi - \
		--out bindings/rrm_manager.go \
		--type ReferralRewardManager \
		--bin $(temp)

		rm $(temp)

.PHONY: \
	event-pool-services \
	bindings \
	binding-rrm \
	clean \
	test \
	lint
