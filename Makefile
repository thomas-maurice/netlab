all: bin binary

# This part of the makefile is adapted from https://gist.github.com/grihabor/4a750b9d82c9aa55d5276bd5503829be
DESCRIBE           := $(shell git tag | sort -V -r | head -n 1)

ifeq ($(DESCRIBE),)
DESCRIBE = v0.0.0
endif

DESCRIBE_PARTS     := $(subst -, ,$(DESCRIBE))

VERSION_TAG        := $(word 1,$(DESCRIBE_PARTS))
COMMITS_SINCE_TAG  := $(word 2,$(DESCRIBE_PARTS))

VERSION            := $(subst v,,$(VERSION_TAG))
VERSION_PARTS      := $(subst ., ,$(VERSION))

MAJOR              := $(word 1,$(VERSION_PARTS))
MINOR              := $(word 2,$(VERSION_PARTS))
MICRO              := $(word 3,$(VERSION_PARTS))

NEXT_MAJOR         := $(shell echo $$(($(MAJOR)+1)))
NEXT_MINOR         := $(shell echo $$(($(MINOR)+1)))
NEXT_MICRO         := $(shell echo $$(($(MICRO)+1)))

_dirty_files       := $(shell git status --untracked-files=no --porcelain | wc -l)
ifeq ($(_dirty_files),0)
DIRTY := false
else
DIRTY := true
endif

HASH               := $(shell git rev-parse --short HEAD)
COMMITS_SINCE_TAG  := $(shell git log $(shell git describe --tags --abbrev=0)..HEAD --oneline | wc -l)
BUILD_USER         := $(shell whoami)


ifeq ($(BUMP),)
BUMP := micro
endif

ifeq ($(MAJOR),)
MAJOR := 0
endif

ifeq ($(MINOR),)
MINOR := 0
endif

ifeq ($(MICRO),)
MICRO := 0
endif

ifeq ($(BUMP),minor)
BUMPED_VERSION_NO_V := $(MAJOR).$(NEXT_MINOR).0
endif
ifeq ($(BUMP),major)
BUMPED_VERSION_NO_V := $(NEXT_MAJOR).0.0
endif
ifeq ($(BUMP),micro)
BUMPED_VERSION_NO_V := $(MAJOR).$(MINOR).$(NEXT_MICRO)
endif

BUMPED_VERSION := v$(BUMPED_VERSION_NO_V)
VERSION_NO_V := $(MAJOR).$(MINOR).$(MICRO)

ifneq ($(COMMITS_SINCE_TAG),0)
VERSION_NO_V := $(VERSION_NO_V)-$(COMMITS_SINCE_TAG)
endif

ifeq ($(DIRTY),true)
VERSION_NO_V := $(VERSION_NO_V)-$(HASH)-dirty-$(BUILD_USER)
endif

VERSION := v$(VERSION_NO_V)

version:
	@echo "Version           : $(VERSION), no v: $(VERSION_NO_V)"
	@echo "Bumped version    : $(BUMPED_VERSION), no v: $(BUMPED_VERSION_NO_V)"
	@echo "Bump              : $(BUMP)"
	@echo "Commits since tag : $(COMMITS_SINCE_TAG)"
	@echo "SHA1 hash         : $(HASH)"

# End of the semver part

GOENV   :=
GOFLAGS := -ldflags \
	"\
	-X 'github.com/thomas-maurice/netlab/pkg/version.Version=$(VERSION)' \
	-X 'github.com/thomas-maurice/netlab/pkg/version.BuildHost=$(shell hostname)' \
	-X 'github.com/thomas-maurice/netlab/pkg/version.BuildTime=$(shell date)' \
	-X 'github.com/thomas-maurice/netlab/pkg/version.BuildHash=$(HASH)' \
	-X 'github.com/thomas-maurice/netlab/pkg/version.OS=$(shell go env GOOS)' \
	-X 'github.com/thomas-maurice/netlab/pkg/version.Arch=$(shell go env GOARCH)' \
	"

BINARY_SUFFIX := $(VERSION)_$(shell go env GOOS)_$(shell go env GOARCH)

clean:
	rm -rf ./bin __pycache__ gowerline.egg-info build dist

.PHONY: bump-version
bump-version:
	if [ git branch | grep \* | cut -f 2 -d \ != "master" ]; then echo "This must be ran from master"; exit 1; fi;
	echo "$(BUMPED_VERSION)" > VERSION
	git add VERSION
	git commit -m "bump version $(VERSION) -> $(BUMPED_VERSION)"
	git tag $(BUMPED_VERSION) -m "bump version $(VERSION) -> $(BUMPED_VERSION)"
	@echo "Don't forget to git push --tags, run make push_tags"

push_tags:
	git push
	git push --tags

autobump: bump-version push_tags

bin:
	if ! [ -d bin ]; then mkdir bin; fi;

binary: bin
	go build -o bin/netlab
