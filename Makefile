LAMBDA_SRC = pkg/*/*.go
LAMBDA_FUNCTIONS = \
	build/api \
	build/enqueueScan \
	build/inspect \
	build/notify \
	build/scanImage \
	build/updateDB

TRIVY_VERSION=0.16.0
TRIVY_URL := https://github.com/aquasecurity/trivy/releases/download/v$(TRIVY_VERSION)/trivy_$(TRIVY_VERSION)_Linux-64bit.tar.gz
TRIVY_BIN=./build/trivy

lambda: $(LAMBDA_FUNCTIONS)

build/api: lambda/api/*.go $(LAMBDA_SRC)
	go build -o build/api ./lambda/api 
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

asset: trivy lambda
	cp build/* /asset-output/
