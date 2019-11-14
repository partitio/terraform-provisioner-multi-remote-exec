PROJECT_NAME := terraform-provisioner-multi-remote-exec

.PHONY: install
install:
	@go install -v ./cmd/$(PROJECT_NAME)

.PHONY: go-test
go-test:
	@go test -v .

.PHONY: test
test: clean go-test install
	@cd tests; \
		echo "Starting test sshd container on port 2222"; \
		docker run --name test-sshd -d --rm --publish=2222:22 sickp/alpine-sshd:7.9-r1; \
		echo "\n***** Executing: terraform init *****"; \
		terraform init; \
		echo "\n***** Executing: terraform plan *****"; \
		terraform plan; \
		echo "\n***** Executing: terraform apply -auto-approve *****"; \
		terraform apply -auto-approve; \
		echo "\n***** Executing: terraform show *****"; \
		terraform show; \
		echo "\n***** Executing: terraform plan *****"; \
		terraform plan; \
		echo "Cleaning ..."; \
		docker stop test-sshd &> /dev/null

.PHONY: clean
clean:
	@rm -rf $GOBIN/terraform-provisioner-multi-remote-exec
	@docker stop test-sshd &> /dev/null ||true
	@rm -rf tests/terraform.tfstate tests/crash.log tests/.terraform.tfstate.lock.info tests/.terraform
