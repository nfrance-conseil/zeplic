#!/bin/sh
#
#Makefile for zeplic
#

GOPATH:=$(shell go env GOPATH)
GOGET=$(shell) go get
GOBUILD=$(shell) go install

PACKAGE1="github.com/mistifyio/go-zfs"
PACKAGE2="github.com/pborman/uuid"
PACKAGE3="github.com/sevlyar/go-daemon"
PACKAGE4="github.com/kardianos/osext"

make:
	@printf "\n:: ZEPLIC ::\n"
	@printf "\nBuilding tree... "
	@if [ ! -d "$(GOPATH)/pkg" ] ; then mkdir -p "$(GOPATH)/pkg" ; fi
	@if [ ! -d "$(GOPATH)/bin" ] ; then mkdir -p "$(GOPATH)/bin" ; fi
	@printf "done!"
	@printf "\nGetting dependencies... "
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE1)" ] ; then $(GOGET) $(PACKAGE1) ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE2)" ] ; then $(GOGET) $(PACKAGE2) ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE3)" ] ; then $(GOGET) $(PACKAGE3) ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE4)" ] ; then $(GOGET) $(PACKAGE4) ; fi
	@printf "done!"
	@printf "\nSetting syslog daemon service... "
	@if ! grep -q \!zeplic "/etc/syslog.conf" ; then printf "\n\!zeplic\nlocal0.*\t\t\t\t\t-/var/log/zeplic.log\n" >> /etc/syslog.conf ; fi
	@printf "done!"
	@printf "\nBuilding... "
	@$(GOBUILD)
	@printf "done!"
	@printf "\nExporting zeplic... "
	@scp $(GOPATH)/bin/zeplic /usr/local/bin/
	@printf "done!\n\n"
	@printf "Remember to config your JSON file: /usr/local/etc/zeplic.d/config.json\n\n"

clean:
	@printf "\n:: ZEPLIC ::\n"
	@printf "\nCleaning dependencies... \c"
	@rm -rf "$(GOPATH)/src/$(PACKAGE4)"
	@rmdir "$(GOPATH)/src/$(PACKAGE4)" 2>/dev/null || :
	@rm -rf "$(GOPATH)/src/$(PACKAGE3)"
	@rmdir "$(GOPATH)/src/$(PACKAGE3)" 2>/dev/null || :
	@rm -rf "$(GOPATH)/src/$(PACKAGE2)"
	@rmdir "$(GOPATH)/src/$(PACKAGE2)" 2>/dev/null || :
	@rm -rf "$(GOPATH)/src/$(PACKAGE1)"
	@rmdir "$(GOPATH)/src/$(PACKAGE1)" 2>/dev/null || :
#	@rm -rf "$(GOPATH)/src/$(PACKAGE0)"
#	@rmdir "$(GOPATH)/src/$(PACKAGE0)" 2>/dev/null || :
	@printf "done!\n\n"
