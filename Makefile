all: build

build:
	CGO_ENABLED=0 go build

# This will trigger a release build in github actions
RELEASE-INCREMENTS:=major minor patch

define release_template =
release-$(1):
	@ \
	NEW_VERSION=$$$$(awk -F. -vOFS=. -vinc=$(1) \
	    '{ \
	        ma=$$$$1; \
			gsub(/^v/,"",ma); \
			mi=$$$$2; \
			pa=$$$$3; \
			if(inc=="major")ma++; \
			if(inc=="minor")mi++; \
			if(inc=="patch")pa++; \
			print "v"ma,mi,pa; \
		}' \
		VERSION \
	) ;\
	echo "===== Creating a release with the version: $$$$NEW_VERSION =====" ;\
	set -x ;\
	git pull ;\
	git pull --tags ;\
	git tag $$$$NEW_VERSION ;\
	git push --tags ;\
	git push
endef

# Three targets are generated: release-patch release-minor release-major
$(foreach increment,$(RELEASE-INCREMENTS),$(eval $(call release_template,$(increment))))
