include ../../../Makefile.defs

BUILDDIR := "$(CURDIR)/cilium-$(VERSION)"
BASEDIR := "$(CURDIR)/../../../"
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

build: clean
	mkdir -p output
	(cd $(BASEDIR) && git bundle create $(CURDIR)/version_$(VERSION) HEAD )
	(cd $(CURDIR) && git clone $(CURDIR)/version_$(VERSION) cilium)
	docker build -t cilium:cilium-bin-deb-$(VERSION) $(CURDIR)
	docker run --rm -v $(CURDIR)/output:/output cilium:cilium-bin-deb-$(VERSION)

clean:
	rm -rf cilium output version_*

test:

	for svc in $(shell ls /tmp/*.service); do \
		cp -f $$svc /tmp/puta/; \
		a=`echo $$svc | sed -n 's/.*\(cilium\-.*\)\.service/\1/p'`;\
		echo $$a;\
		echo "$$svc"; \
	done

.PHONY: force build clean
force :;

