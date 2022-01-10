#!/bin/bash

mkdir -p ./libs/

go get golang.org/x/mobile/bind

gomobile bind -v -o ./libs/astral.aar -target=android ../node/
