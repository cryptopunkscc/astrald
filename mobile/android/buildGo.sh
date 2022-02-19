#!/bin/bash

mkdir -p ./build/

go get golang.org/x/mobile/bind

gomobile bind -v -o ./build/astral.aar -target=android ./node/
