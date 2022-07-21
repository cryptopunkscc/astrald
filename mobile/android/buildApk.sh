#!/bin/bash

test -f ./build/astral.aar || ./buildGo.sh

./gradlew :astral:app:assembleDebug

cp ./app/build/outputs/apk/debug/app-debug.apk ./build/astral-android.apk

echo "$(pwd)/build/astral-android.apk"
