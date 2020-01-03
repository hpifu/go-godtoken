binary=godtoken
dockeruser=hatlonely
gituser=hpifu
repository=go-godtoken
version=$(shell git describe --tags)

export GOPATH=$(shell pwd)/../../../../
export PATH:=${PATH}:${GOPATH}/bin:$(shell pwd)/third/go/bin:$(shell pwd)/third/protobuf/bin:$(shell pwd)/third/cloc-1.76:$(shell pwd)/third/redis-3.2.8/src

.PHONY: all
all: third vendor output test stat

.PHONY: deploy
deploy:
	mkdir -p /var/docker/${repository}/log
	docker stack deploy -c stack.yml ${repository}

.PHONY: deploytest
deploytest:
	if [ -z "$(shell docker network ls --filter name=testnet -q)" ]; then \
		docker network create -d bridge testnet; \
	fi
	if [ ! -z "$(shell docker ps -a --filter name=test-go-godtoken -q)" ]; then \
		docker stop test-go-godtoken && docker rm test-go-godtoken; \
	fi
	docker run --name test-go-godtoken --hostname test-go-godtoken --network testnet -d \
		-e GODTOKEN_REDIS_ADDR="test-redis:6379" \
		-e GODTOKEN_SERVICE_PORT="17060" \
		-e GODTOKEN_ES_URI="http://test-elasticsearch:9200" \
		${dockeruser}/${repository}:${version}

.PHONY: remove
remove:
	docker stack rm ${repository}

.PHONY: push
push:
	docker push ${dockeruser}/${repository}:${version}

.PHONY: buildenv
buildenv:
	if [ -z "$(shell docker network ls --filter name=testnet -q)" ]; then \
		docker network create -d bridge testnet; \
	fi
	if [ -z "$(shell docker ps -a --filter name=go-build-env -q)" ]; then \
		docker run --name go-build-env --network testnet -d hatlonely/go-env:v1.1.0 tail -f /dev/null; \
	fi
	if [ -z "$(shell docker ps -a --filter name=test-redis -q)" ]; then \
		docker run --name test-redis --hostname test-redis --network testnet -d redis:5.0.5-alpine; \
	fi

.PHONY: cleanbuildenv
cleanbuildenv:
	if [ -z "$(shell docker ps -a --filter name=test-redis -q)" ]; then \
		docker run --name test-redis --hostname test-redis --network testnet -d redis:5.0.5-alpine; \
	fi
	if [ ! -z "$(shell docker ps -a --filter name=go-build-env -q)" ]; then \
		docker stop go-build-env  && docker rm go-build-env; \
	fi
	if [ ! -z "$(shell docker network ls --filter name=testnet -q)" ]; then \
		docker network rm testnet; \
	fi

.PHONY: image
image: buildenv
	docker exec go-build-env rm -rf /data/src/${gituser}/${repository}
	docker exec go-build-env mkdir -p /data/src/${gituser}/${repository}
	docker cp . go-build-env:/data/src/${gituser}/${repository}
	docker exec go-build-env bash -c "cd /data/src/${gituser}/${repository} && make output"
	mkdir -p docker/
	docker cp go-build-env:/data/src/${gituser}/${repository}/output/${repository} docker/
	docker build --tag=hatlonely/${repository}:${version} .
	cat stack.tpl.yml | sed 's/\$${version}/${version}/g' | sed 's/\$${repository}/${repository}/g' > stack.yml

.PHONY: dockertest
dockertest: buildenv
	# docker exec go-build-env rm -rf /data/src/${gituser}/${repository}
	# docker exec go-build-env mkdir -p /data/src/${gituser}/${repository}
	# docker cp . go-build-env:/data/src/${gituser}/${repository}
	docker exec go-build-env bash -c "cd /data/src/${gituser}/${repository} && make test"

.PHONY: dockerbehave
dockerbehave: buildenv
	# docker exec go-build-env rm -rf /data/src/${gituser}/${repository}
	# docker exec go-build-env mkdir -p /data/src/${gituser}/${repository}
	# docker cp . go-build-env:/data/src/${gituser}/${repository}
	docker exec go-build-env bash -c "cd /data/src/${gituser}/${repository} && make behave"

output: cmd/*/*.go internal/*/*.go scripts/version.sh Makefile api vendor
	@echo "compile"
	@go build -ldflags "-X 'main.AppVersion=`sh scripts/version.sh`'" cmd/${binary}/main.go && \
	mkdir -p output/${repository}/bin && mv main output/${repository}/bin/${binary} && \
	mkdir -p output/${repository}/configs && cp configs/${binary}/* output/${repository}/configs && \
	mkdir -p output/${repository}/log
	@go build -ldflags "-X 'main.AppVersion=`sh scripts/version.sh`'" cmd/${binary}-cli/main.go && \
	mkdir -p output/${repository}/bin && mv main output/${repository}/bin/${binary}-cli

vendor: go.mod
	@echo "install golang dependency"
	go mod vendor

.PHONY: api
api: api/godtoken.pb.go Makefile
	cd api && python3 -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. godtoken.proto

%_easyjson.go: %.go
	easyjson $<

%.pb.go: %.proto
	protoc --go_out=plugins=grpc:. $<

.PHONY: test
test: api vendor
	@echo "Run unit tests"
	cd internal && go test -cover ./...

.PHONY: behave
behave: output api
	cp api/*.py .
	behave features

.PHONY: stat
stat: cloc gocyclo
	@echo "code statistics"
	cloc internal Makefile --by-file
	@echo "circle complexity statistics"
	gocyclo internal
	@gocyclo internal | awk '{sum+=$$1}END{printf("complexity: %s", sum)}'

.PHONY: clean
clean:
	rm -rf output
	rm -rf api/*.py
	rm -rf *_pb2_grpc.py *_pb2.py __pycache__

.PHONY: deep_clean
deep_clean:
	rm -rf output vendor third

third: protoc golang cloc gocyclo easyjson

.PHONY: protoc
protoc: golang
	@hash protoc 2>/dev/null || { \
		echo "install protobuf codegen tool protoc" && \
		mkdir -p third && cd third && \
		wget https://github.com/google/protobuf/releases/download/v3.2.0/protobuf-cpp-3.2.0.tar.gz && \
		tar -xzvf protobuf-cpp-3.2.0.tar.gz && \
		cd protobuf-3.2.0 && \
		./configure --prefix=`pwd`/../protobuf && \
		make -j8 && \
		make install && \
		cd ../.. && \
		protoc --version; \
	}
	@hash protoc-gen-go 2>/dev/null || { \
		echo "install protobuf golang plugin protoc-gen-go" && \
		go get -u github.com/golang/protobuf/{proto,protoc-gen-go}; \
	}

.PHONY: golang
golang:
	@hash go 2>/dev/null || { \
		echo "install go1.9" && \
		mkdir -p third && cd third && \
		wget https://dl.google.com/go/go1.9.linux-amd64.tar.gz && \
    	tar -xzvf go1.9.linux-amd64.tar.gz && \
		cd .. && \
		go version; \
	}

.PHONY: cloc
cloc:
	@hash cloc 2>/dev/null || { \
		echo "install cloc" && \
		mkdir -p third && cd third && \
		wget https://github.com/AlDanial/cloc/archive/v1.76.zip && \
		unzip v1.76.zip; \
	}

.PHONY: gocyclo
gocyclo: golang
	@hash gocyclo 2>/dev/null || { \
		echo "install gocyclo" && \
		go get -u github.com/fzipp/gocyclo; \
	}

.PHONY: easyjson
easyjson: golang
	@hash easyjson 2>/dev/null || { \
		echo "install easyjson" && \
		go get -u github.com/mailru/easyjson/...; \
	}

.PHONY: redis
redis:
	@hash redis 2>/dev/null || { \
		echo "install redis" && \
		mkdir -p third && cd third && \
		wget http://download.redis.io/releases/redis-3.2.8.tar.gz && \
		tar -xzvf redis-3.2.8.tar.gz && \
		cd redis-3.2.8 && \
		make -j8; \
	}
