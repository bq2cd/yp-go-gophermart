MISE := $(shell which mise 2>/dev/null)

all:
ifeq ($(MISE),)
	@echo "Looks like 'mise' binary is not installed or not in the PATH."
	@echo
	@echo "Please visit https://mise.jdx.dev/getting-started.html to install it first."
	@echo
else
	mise run setup
endif
