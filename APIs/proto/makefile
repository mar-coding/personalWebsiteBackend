errors:
	protoc --proto_path=components --go_out=. components/errors/v1/errors.proto
	$(MAKE) move-to-apis-gen

microservice:
	protoc --proto_path=components --proto_path=includes --go_out=. --go-grpc_out=. components/microservice/v1/*.proto
	$(MAKE) move-to-apis-gen

list:
	protoc --proto_path=components --go_out=. --go-grpc_out=. components/list/v1/*.proto
	$(MAKE) move-to-apis-gen

website:
	protoc --proto_path=services/website/v1 --proto_path=. --proto_path=includes/googleapis --proto_path=includes/gateway --proto_path=includes/validator --proto_path=includes --go-grpc_out=. --go_out=. --grpc-gateway_out=logtostderr=true:. --validate_out=lang=go:.  services/website/v1/*.proto
	# swagger
	protoc --proto_path=services/website/v1 --proto_path=. --proto_path=includes/googleapis --proto_path=includes/gateway --proto_path=includes/validator --proto_path=includes --openapiv2_out=allow_merge=true:api --openapiv2_opt use_go_templates=true services/website/v1/*.proto
	$(MAKE) move-to-apis-gen



submodule:
	git submodule update --init --recursive

proto-tools:
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@printf "download protoc from : https://github.com/protocolbuffers/protobuf/releases or for debian/ubuntu apt install -y protobuf-compiler \n"
	@printf "in case of 'protoc-gen-go-grpc: program not found or is not executable' error you most run: 'export GOPATH=$HOME/go' and 'export PATH=$PATH:$GOPATH/bin' "

move-to-apis-gen:
	cp -r ./github.com/mar-coding/personalWebsiteBackend/APIs/proto-gen/* ../proto-gen
	rm -r ./github.com
	cp api/apidocs.swagger.json ../