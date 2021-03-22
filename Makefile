ROOT := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
ASSET_OUTPUT = /asset-output
LAMBDA_SRC = pkg/*/*.go
LAMBDA_FUNCTIONS = \
	build/apiHandler \
	build/enqueueScan \
	build/inspect \
	build/notify \
	build/scanImage \
	build/updateDB

TRIVY_VERSION=0.16.0
TRIVY_URL := https://github.com/aquasecurity/trivy/releases/download/v$(TRIVY_VERSION)/trivy_$(TRIVY_VERSION)_Linux-64bit.tar.gz
TRIVY_BIN=./build/trivy

lambda: $(LAMBDA_FUNCTIONS)

build/apiHandler: lambda/apiHandler/*.go $(LAMBDA_SRC)
	go build -o build/apiHandler ./lambda/apiHandler
build/enqueueScan: lambda/enqueueScan/*.go $(LAMBDA_SRC)
	go build -o build/enqueueScan ./lambda/enqueueScan
build/inspect: lambda/inspect/*.go $(LAMBDA_SRC)
	go build -o build/inspect ./lambda/inspect
build/notify: lambda/notify/*.go $(LAMBDA_SRC)
	go build -o build/notify ./lambda/notify
build/scanImage: lambda/scanImage/*.go $(LAMBDA_SRC)
	go build -o build/scanImage ./lambda/scanImage
build/updateDB: lambda/updateDB/*.go $(LAMBDA_SRC)
	go build -o build/updateDB ./lambda/updateDB

trivy: $(TRIVY_BIN)

$(TRIVY_BIN):
	$(eval TRIVY_TMPDIR := $(shell mktemp -d))
	mkdir -p build
	curl -o $(TRIVY_TMPDIR)/trivy.tar.gz -s -L $(TRIVY_URL)
	tar -C $(TRIVY_TMPDIR) -xzf $(TRIVY_TMPDIR)/trivy.tar.gz
	mv $(TRIVY_TMPDIR)/trivy $(TRIVY_BIN)
	rm -r $(TRIVY_TMPDIR)

FRONTEND_DIR = $(ROOT)/frontend
BUNDLE_JS = $(FRONTEND_DIR)/dist/bundle.js
JS_SRC = $(FRONTEND_DIR)/src/js/*.tsx

$(BUNDLE_JS): $(JS_SRC)
	cd $(FRONTEND_DIR) && npm i && npm exec webpack && cd $(ROOT)

js: $(BUNDLE_JS)

asset: trivy lambda js
	cp build/* $(ASSET_OUTPUT)
	cp -r frontend/dist $(ASSET_OUTPUT)/assets
