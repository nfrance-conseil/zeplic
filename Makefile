#!/bin/sh
#
#Makefile for zeplic
#

GOPATH:=$(shell go env GOPATH)
GOGET=$(shell) go get
GOBUILD=$(shell) go install

PACKAGE0="github.com/nfrance-conseil"
PACKAGE1="github.com/mistifyio"
PACKAGE2="github.com/pborman"
PACKAGE3="github.com/sevlyar"
PACKAGE4="github.com/kardianos"

make:
	@printf "\n:: ZEPLIC ::\n"
	@printf "\nBuilding tree... "
	@if [ ! -d "$(GOPATH)/pkg" ] ; then sudo mkdir -p "$(GOPATH)/pkg" ; fi
	@if [ ! -d "$(GOPATH)/bin" ] ; then sudo mkdir -p "$(GOPATH)/bin" ; fi
	@printf "done!"
	@printf "\nGetting dependencies... "
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE1)" ] ; then sudo $(GOGET) $(PACKAGE1)/go-zfs ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE2)" ] ; then sudo $(GOGET) $(PACKAGE2)/uuid ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE3)" ] ; then sudo $(GOGET) $(PACKAGE3)/go-daemon ; fi
	@if [ ! -d "$(GOPATH)/src/$(PACKAGE4)" ] ; then sudo $(GOGET) $(PACKAGE4)/osext ; fi
	@printf "done!"
	@printf "\nSetting syslog daemon service... "
	@if [ -e "/etc/rsyslog.conf" ] && ! grep -q \!zeplic "/etc/rsyslog.conf" ; then sudo printf "\n\!zeplic\nlocal0.*\t\t\t\t\t-/var/log/zeplic.log\n" >> /etc/rsyslog.conf ; fi
	@if [ -e "/etc/syslog.conf" ] && ! grep -q \!zeplic "/etc/syslog.conf" ; then sudo printf "\n\!zeplic\nlocal0.*\t\t\t\t\t-/var/log/zeplic.log\n" >> /etc/syslog.conf ; fi	
	@printf "done!"
	@printf "\nBuilding... "
	@$(GOBUILD)
	@printf "done!"
	@printf "\nInstalling zeplic... "
	@sudo install -m 755 -o root -g wheel -b $(GOPATH)/bin/zeplic /usr/local/bin
	@printf "done!\n\n"
	@printf "Remember to config your JSON file: /usr/local/etc/zeplic.d/config.json\n\n"

clean:
	@printf "\n:: ZEPLIC ::\n"
	@printf "\nCleaning dependencies... "
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE4)/osext"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE4)" 2>/dev/null || :
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE3)/go-daemon"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE3)" 2>/dev/null || :
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE2)/uuid"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE2)" 2>/dev/null || :
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE1)/go-zfs"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE1)" 2>/dev/null || :
	@sudo rm -rf "$(GOPATH)/src/$(PACKAGE0)/zeplic"
	@sudo rmdir "$(GOPATH)/src/$(PACKAGE0)" 2>/dev/null || :
	@printf "done!\n\n"
