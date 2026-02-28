PWD= $(shell pwd)
BACK = $(PWD)/back
FRONT = $(PWD)/front

.PHONY: back front mkdir1 mkdir-front back-dependcies front-dependcies

all: back front

back-dependcies:
	cd $(BACK)/src && go mod tidy

front-dependcies:
	cd $(FRONT) && IBM_TELEMETRY_DISABLED='true';npm install

mkdir1:
	mkdir -p $(PWD)/artifacts

mkdir-front:
	mkdir -p $(PWD)/artifacts/FrontEndBuild


back: mkdir1 back-dependcies
	cd $(BACK) && make server && cp $(BACK)/build/wts $(PWD)/artifacts/wts

front: mkdir-front front-dependcies
	cd $(FRONT) && IBM_TELEMETRY_DISABLED='true';npm run build && cp -r  $(FRONT)/build/* $(PWD)/artifacts/FrontEndBuild
