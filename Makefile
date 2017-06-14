#Based on https://github.com/cloudflare/hellogopher - v1.1 - MIT License
#
## Copyright (c) 2017 Cloudflare
#
## Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

IMPORT_PATH := github.com/nfrance-conseil/zeplic
CHECK := 0
i := 0

.PHONY: all
all: build

.PHONY: build
build: .GOPATH/.ok
	$Q printf "\nLet's build zeplic..."
	$Q printf "\nGetting dependencies... "
	$Q go get $(if $V,-v) github.com/mistifyio/go-zfs
	$Q go get $(if $V,-v) github.com/pborman/uuid
	$Q go get $(if $V,-v) github.com/sevlyar/go-daemon
	$Q go install $(if $V,-v) $(COMPILE_FLAGS) $(IMPORT_PATH)
	$Q printf "done!"
	$Q printf "\n\nBUILD! To install, type: sudo make install\n\n"

### Code not in the repository root? Another binary? Add to the path like this.
# .PHONY: otherbin
# otherbin: .GOPATH/.ok
#	$Q go install $(if $V,-v) $(COMPILE_FLAGS) $(IMPORT_PATH)/cmd/otherbin

.PHONY: clean test list cover format

clean:
	$Q rm -rf bin .GOPATH

test: .GOPATH/.ok
	$Q go test $(if $V,-v) -i -race $(allpackages) # install -race libs to speed up next run
ifndef CI
	$Q go vet $(allpackages)
	$Q GODEBUG=cgocheck=2 go test -race $(allpackages)
else
	$Q ( go vet $(allpackages); echo $$? ) | \
	    tee .GOPATH/test/vet.txt | sed '$$ d'; exit $$(tail -1 .GOPATH/test/vet.txt)
	$Q ( GODEBUG=cgocheck=2 go test -v -race $(allpackages); echo $$? ) | \
	    tee .GOPATH/test/output.txt | sed '$$ d'; exit $$(tail -1 .GOPATH/test/output.txt)
endif

list: .GOPATH/.ok
	@echo ""
	@echo $(allpackages)
	@echo ""

cover: bin/gocovmerge .GOPATH/.ok
	@echo "NOTE: make cover does not exit 1 on failure, don't use it to check for tests success!"
	$Q rm -f .GOPATH/cover/*.out .GOPATH/cover/all.merged
	$(if $V,@echo "-- go test -coverpkg=./... -coverprofile=.GOPATH/cover/... ./...")
	@for MOD in $(allpackages); do \
		go test -coverpkg=`echo $(allpackages)|tr " " ","` \
			-coverprofile=.GOPATH/cover/unit-`echo $$MOD|tr "/" "_"`.out \
			$$MOD 2>&1 | grep -v "no packages being tested depend on"; \
	done
	$Q ./bin/gocovmerge .GOPATH/cover/*.out > .GOPATH/cover/all.merged
ifndef CI
	$Q go tool cover -html .GOPATH/cover/all.merged
else
	$Q go tool cover -html .GOPATH/cover/all.merged -o .GOPATH/cover/all.html
endif
	@echo ""
	@echo "=====> Total test coverage: <====="
	@echo ""
	$Q go tool cover -func .GOPATH/cover/all.merged

format: bin/goimports .GOPATH/.ok
	$Q find .GOPATH/src/$(IMPORT_PATH)/ -iname \*.go | grep -v \
	    -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)) | xargs ./bin/goimports -w

.PHONY: install
install:
	$Q printf "\nLet's install zeplic..."
	$Q printf "\nInstalling zeplic in your BIN directory... "
	$Q install $(if $V,-v) -m 755 .GOPATH/bin/zeplic $(BINDIR) 
	$Q printf "done!"
	$Q printf "\nCreating a sample JSON file in $(SYSCONFDIR)/zeplic/... "
	$Q mkdir -p $(SYSCONFDIR)/zeplic
	$Q install $(if $V,-v) -m 644 samples/config.json.sample $(SYSCONFDIR)/zeplic
	$Q printf "done!"
	$Q printf "\nConfiguring your syslog daemon service... "
	$Q $(eval CHECK := $(shell if grep -q \!zeplic "$(SYSLOG)" ; then echo $$(( $(CHECK) +1 )) ; else echo $(CHECK) ; fi ) )
	$Q if [ $(CHECK) -eq 0 ] ; then \
		i=$(i) ; \
		while [ $${i} -le 7 ] ; do \
		if grep -q local$$i.* "$(SYSLOG)" ; then i=`expr $$i + 1` ; else printf "\n\!zeplic\nlocal$$i.*\t\t\t\t\t-/var/log/zeplic.log\n" >> $(SYSLOG) && break ; fi ; \
		done ; \
		true ; \
	   fi
	$Q printf "done!"
	$Q printf "\n\nINSTALL! Remember to config your JSON file.\n\n"

##### =====> Internals <===== #####

VERSION          := $(shell git describe --tags --always --dirty="-dev")
DATE             := $(shell date -u '+%Y-%m-%d-%H%M UTC')
OS 		 := $(shell uname)
ifeq ($(OS),FreeBSD)
SYSCONFDIR 	 := /usr/local/etc
BINDIR           := /usr/local/bin
SYSLOG		 := /etc/syslog.conf
else
SYSCONFDIR       := /etc
BINDIR           := /usr/bin
SYSLOG		 := /etc/rsyslog.conf
endif
COMPILE_FLAGS    := -ldflags='-X "main.Version=$(VERSION)" -X "main.BuildTime=$(DATE)" -X "github.com/nfrance-conseil/zeplic/config.ConfigFilePath=$(SYSCONFDIR)/zeplic/config.json" -X "github.com/nfrance-conseil/zeplic/config.SyslogFilePath=$(SYSLOG)"'

# cd into the GOPATH to workaround ./... not following symlinks
_allpackages = $(shell ( cd $(CURDIR)/.GOPATH/src/$(IMPORT_PATH) && \
    GOPATH=$(CURDIR)/.GOPATH go list ./... 2>&1 1>&3 | \
    grep -v -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)) 1>&2 ) 3>&1 | \
    grep -v -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)))

# memoize allpackages, so that it's executed only once and only if used
allpackages = $(if $(__allpackages),,$(eval __allpackages := $$(_allpackages)))$(__allpackages)

export GOPATH := $(CURDIR)/.GOPATH
unexport GOBIN

Q := $(if $V,,@)

.GOPATH/.ok:
	$Q mkdir -p "$(dir .GOPATH/src/$(IMPORT_PATH))"
	$Q ln -s ../../../.. ".GOPATH/src/$(IMPORT_PATH)"
	$Q mkdir -p .GOPATH/test .GOPATH/cover
	$Q mkdir -p .GOPATH/bin
	$Q touch $@
