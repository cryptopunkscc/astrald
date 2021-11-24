#!/bin/bash

mkdir -p ./libs/

gomobile bind -v -o ./libs/astral.aar -target=android ../node/
