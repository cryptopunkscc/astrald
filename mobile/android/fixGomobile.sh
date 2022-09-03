#!/bin/bash

# Make sure the go mobile is installed
go install golang.org/x/mobile/cmd/gomobile@latest

# Android studio doesn't source ~/bashrc for gradle tasks.
# Link required binaries to avoid errors while compiling go mobile through android studio.
ln -f -s ~/go/bin/gomobile ~/.local/bin/
ln -f -s ~/go/bin/gobind ~/.local/bin/
