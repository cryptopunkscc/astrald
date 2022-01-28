#!/bin/bash

test -f ./libs/astral.aar || ./buildGo.sh

./gradlew node:app:assembleDebug

cp ./node/app/build/outputs/apk/debug/app-debug.apk ./libs/astral-android.apk

echo "$(pwd)/libs/astral-android.apk"
