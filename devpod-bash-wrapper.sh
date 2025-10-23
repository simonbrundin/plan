#!/bin/bash
# Wrapper script to run devpod with bash as the shell
# This ensures compatibility when nushell is the default shell

export SHELL=/bin/bash
exec devpod "$@"