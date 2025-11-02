legoVersion=4.24.0
legoArchive=lego.tar.gz

test: test_nginx test_apache

test_nginx:
	docker run --volume="$(shell pwd):/opt/r2dtools" sslbot-nginx-tests
test_apache:
	docker run --volume="$(shell pwd):/opt/r2dtools" sslbot-apache-tests

build_agent:
	go build -tags prod -ldflags="-s -w -X 'main.Version=${version}'" -o ./build/sslbot -v cmd/main.go

build_lego:
	wget "https://github.com/go-acme/lego/releases/download/v${legoVersion}/lego_v$(legoVersion)_linux_amd64.tar.gz" -O $(legoArchive); \
	tar -xvzf $(legoArchive) -C build lego; \
	rm $(legoArchive)

build: build_agent build_lego
	cp LICENSE build/

build_test: build_apache_test build_nginx_test

build_nginx_test:
	docker build -f Dockerfile.nginx -t sslbot-nginx-tests . 
build_apache_test:
	docker build -f Dockerfile.apache -t sslbot-apache-tests . 
clean:
	cd build; \
	rm -rf config; \
	rm -f lego sslbot LICENSE

serve:
	go run cmd/main.go serve

.PHONY: test
