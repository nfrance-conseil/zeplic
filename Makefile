#!/bin/sh
#
#Makefile for zeplic
#

GOPATH:=$(shell go env GOPATH)
GOGET=$(shell) go get -v
#GOBUILD=$(shell) go install

PACKAGE1="github.com/mistifyio/go-zfs"
PACKAGE2="github.com/pborman/uuid"
PACKAGE3="github.com/sevlyar/go-daemon"
PACKAGE4="github.com/kardianos/osext"

make:
	@printf "\n:: ZEPLIC ::\n"
	@printf "\nBuilding tree... "
	@if [ ! -d "$(GOPATH)/pkg" ] ; then sudo mkdir -p "$(GOPATH)/pkg" ; fi
	@if [ ! -d "$(GOPATH)/bin" ] ; then sudo mkdir -p "$(GOPATH)/bin" ; fi
	@printf "done!"
	@printf "\nGetting dependencies... "
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE1)" ] ; then $(GOGET) $(PACKAGE1) ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE2)" ] ; then $(GOGET) $(PACKAGE2) ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE3)" ] ; then $(GOGET) $(PACKAGE3) ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE4)" ] ; then $(GOGET) $(PACKAGE4) ; fi
	@printf "done!"
	@printf "\nSetting syslog daemon service... "
	@if [ -e "/etc/rsyslog.conf" ] && ! grep -q \!zeplic "/etc/rsyslog.conf" ; then sudo printf "\n\!zeplic\nlocal0.*\t\t\t\t\t-/var/log/zeplic.log\n" >> /etc/rsyslog.conf ; fi
	@if [ -e "/etc/syslog.conf" ] && ! grep -q \!zeplic "/etc/syslog.conf" ; then sudo printf "\n\!zeplic\nlocal0.*\t\t\t\t\t-/var/log/zeplic.log\n" >> /etc/syslog.conf ; fi	
	@printf "done!"
	@printf "\nBuilding... "
#	@$(GOBUILD)
	@printf "done!"
	@printf "\nExporting your \$$\GOBIN... "
	@if [ -e "/root/.bashrc" ] && ! grep -q "$(GOPATH)/bin" "/root/.bashrc" ; then sudo printf "export PATH=\$$\PATH:$(GOPATH)/bin" >> ~/.bashrc ; fi
	@if [ -e "/root/.cshrc" ] && ! grep -q "$(GOPATH)/bin" "/root/.cshrc" ; then sudo printf "setenv PATH \"\$$\PATH\":$(GOPATH)/bin" >> ~/.cshrc ; fi
	@printf "done!\n\n"
	@printf "Remember to config your JSON file: /usr/local/etc/zeplic.d/config.json\n\n"

clean:
	@printf "\n:: ZEPLIC ::\n"
	@printf "\nCleaning dependencies... \c"
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE4)"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE4)" 2>/dev/null || :
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE3)"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE3)" 2>/dev/null || :
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE2)"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE2)" 2>/dev/null || :
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE1)"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE1)" 2>/dev/null || :
#	@rm -rf "$(GOPATH)/src/$(PACKAGE0)"
#	@rmdir "$(GOPATH)/src/$(PACKAGE0)" 2>/dev/null || :
	@printf "done!\n\n"
