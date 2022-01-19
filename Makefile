ROOT := td-redis-operator
#TARGETS := operator admin
#TARGETS := admin
TARGETS := operator
REGISTRY := tongduncloud
PROJECT := db
VERSION := `date +%Y%m%d`-`git rev-parse --short=11 HEAD`
LDFLAGS := `./hack/version.sh`

.PHONY: codegen compile build push deploy

codegen:
	go generate -v ./...

build: codegen
	rm -rf _output
	mkdir _output
	@for target in $(TARGETS); do                                       \
		go build                                                        \
			-v                                                          \
			--ldflags "$(LDFLAGS)"                                      \
			-o ./_output/$${target}                                     \
		./cmd/$${target};                                               \
	done

container: codegen
	rm -rf _output
	mkdir _output
	@for target in $(TARGETS); do                                            \
		docker run                                                           \
			--rm                                                             \
			-w /go/src/$(ROOT)                                               \
			-v $(PWD):/go/src/$(ROOT)                                        \
			-v $(GOCACHE):/go/.cache                                         \
			-v $(GOPATH)/pkg/mod:/go/pkg/mod                                 \
			-e GO111MODULE=on                                                \
			-e GOCACHE=/go/.cache                                            \
			-e GOPROXY=https://goproxy.cn                                    \
			golang:1.13.12-alpine3.12                                          \
			go build                                                         \
				-o _output/$${target}                                        \
				-v                                                           \
				--ldflags "$(LDFLAGS)"                                       \
				./cmd/$${target};                                            \
		docker build                                                         \
			-t $(REGISTRY)/$(PROJECT)/td-redis-$${target}:$(VERSION)      \
			-f $(PWD)/build/$${target}/Dockerfile .;                         \
	done

push:
	@for target in $(TARGETS); do                                       \
		docker push                                                     \
			$(REGISTRY)/$(PROJECT)/td-redis-$${target}:$(VERSION);   \
	done

deploy:
	kubectl apply -f $(PWD)/deploy/crd.yaml
	@for target in $(TARGETS); do                      \
		cat $(PWD)/deploy/$${target}/$${target}.yaml | \
			VERSION=$(VERSION) envsubst |              \
			kubectl apply -f -;                        \
	done

