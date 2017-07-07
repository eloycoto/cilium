include ../../../Makefile.defs

BUILDDIR := "$(CURDIR)/cilium-$(VERSION)"

build: clean
	mkdir -p $(BUILDDIR)
	git archive master |  gzip > $(BUILDDIR)/cilium-$(VERSION).tar.gz
	# @$(CURDIR)/../cp-dirs.sh $(BUILDDIR)
	# find $(BUILDDIR) -name ".*" -exec $(RM) -rf {} \;
	# cp -R $(CURDIR)/cfg/* $(BUILDDIR)
	# $(MAKE) -C $(BUILDDIR) clean
	# tar -czf cilium-$(VERSION).tar.gz cilium-$(VERSION)/
	# mv cilium-$(VERSION).tar.gz cilium-$(VERSION)
	# docker build -t cilium:cilium-bin-deb-$(VERSION) $(BUILDDIR)
	# docker run --rm cilium:cilium-bin-deb-$(VERSION) bash -c 'cd .. && tar -c cilium_$(VERSION)*' | tar -xvC .

clean:
	ls -d ./* | grep -vE Makefile\|cfg | xargs $(RM) -rf

.PHONY: force build clean
force :;

