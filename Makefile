TRAVIS_HOME=${GOPATH}/src/${TRAVIS_GO_IMPORT_PATH}
FABRIC_HOME=${GOPATH}/src/github.com/hyperledger/fabric
SDK_HOME=${TRAVIS_HOME}/test/fabric-sdk-go

FABRIC_TAG=v1.4.3
SDK_TAG=v1.0.0-beta1

prepare:
	@docker pull golang:latest
	@docker pull hyperledger/fabric-tools:latest
	@docker pull hyperledger/fabric-baseos:amd64-0.4.15
	# Install golang dep
	@curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	# Install fabric-samples
	@cd $(HOME) && curl -sSL http://bit.ly/2ysbOFE | bash -s -- '-d'

fetch_code:
	@git clone --single-branch --branch $(FABRIC_TAG) https://github.com/hyperledger/fabric.git $(FABRIC_HOME)
	@git clone --single-branch --branch $(SDK_TAG) https://github.com/hyperledger/fabric-sdk-go $(SDK_HOME)

patch_code:
	@cd $(FABRIC_HOME) && git apply $(TRAVIS_HOME)/fabric.patch
	@cd $(SDK_HOME) && git apply $(TRAVIS_HOME)/sdk.patch

install_deps:
	@cd $(FABRIC_HOME) && dep ensure -v && make protos

build_images:
	@cd $(FABRIC_HOME) && make peer-docker && make orderer-docker

run_byfn:
	@cd $(HOME)/fabric-samples/first-network && echo y | ./byfn.sh up

run_test:
	@docker run -t \
		-v $(GOPATH):/go \
		-v $(HOME)/fabric-samples/first-network/crypto-config:/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto \
		-w /go/src/$(TRAVIS_GO_IMPORT_PATH)/test \
		--network=net_byfn \
		golang:latest bash -c \
		'go mod download 2>&1 | awk "!/^go: (finding|downloading|extracting)/" && go test -v -failfast ./main_test.go'

publish_images:
	@docker tag hyperledger/fabric-peer ilyapt/patched-fabric-peer
	@docker tag hyperledger/fabric-orderer ilyapt/patched-fabric-orderer
	@echo "${DOCKER_PASSWD}" | docker login -u ilyapt --password-stdin
	@docker push ilyapt/patched-fabric-peer && docker push ilyapt/patched-fabric-orderer

publish_sdk:
	# Prepare ssh
	@mkdir -p $(HOME)/.ssh && echo "$(DEPLOY_SSH_KEY)" | base64 -d -w 0 > $(HOME)/.ssh/id_rsa
	@chmod 400 $(HOME)/.ssh/id_rsa
	@ssh-keyscan -t rsa github.com > $(HOME)/.ssh/known_hosts

	# Download and move repo, 
	@cd $(HOME) && git clone -n git@github.com:ilyapt/fabric-sdk-go-patched.git
	@cd $(TRAVIS_HOME)/test/fabric-sdk-go && \
		rm -rf .git && mv $(HOME)/fabric-sdk-go-patched/.git . && \
		git reset README.md && git checkout -- README.md && \
		git config user.name "Travis-CI" && git add . && \
		git commit -m "based on https://github.com/ilyapt/fabric-certstore/commit/$(TRAVIS_COMMIT)" && \
		git push origin master

store_logs:
	@mkdir -p $(HOME)/logs
	@for x in `docker ps -a --format '{{.Names}}'`; do docker logs $(x) > $(HOME)/logs/$(x).log 2>&1; done
	@cd $(HOME)/logs/ && tar cfz logs.tgz *.log
	@rm -rf $(HOME)/logs/*.log
