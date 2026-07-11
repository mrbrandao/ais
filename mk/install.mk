PREFIX  ?= /usr/local
BINDIR  := $(PREFIX)/bin

.PHONY: install uninstall

install: build ## - install ais to $(BINDIR)
	install -Dm755 bin/ais \
		$(DESTDIR)$(BINDIR)/ais
	@echo "Installed: $(DESTDIR)$(BINDIR)/ais"

uninstall: ## - remove ais from $(BINDIR)
	rm -f $(DESTDIR)$(BINDIR)/ais
	@echo "Removed: $(DESTDIR)$(BINDIR)/ais"
