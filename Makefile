ROOT_GOPATH=$(shell pwd)
$(info $(ROOT_GOPATH))

#ROOT_GOPATH=$(cd $(dirname $0) && pwd)

VENDOR_GOPATH=$(ROOT_GOPATH)/vendor

#GOPATH=$(VENDOR_GOPATH)
export GOPATH=$(VENDOR_GOPATH)
$(info $(VENDOR_GOPATH))

install:
	mkdir -p $(VENDOR_GOPATH)
	sh ./goapp.sh

clean:
	#rm -rf $(VENDOR_GOPATH)

run:
	GOPATH=$(VENDOR_GOPATH) go run Main.go

clean-run:
	go serve --clear_datastore ./src

test:
	GOPATH=$(VENDOR_GOPATH):$(ROOT_GOPATH) goapp test -v ./src/...

deploy:
	goapp deploy ./src