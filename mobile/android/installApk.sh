#!/bin/bash

test -f ./build/astral-android.apk || ./buildApk.sh

adb install build/astral-android.apk
adb shell am start -n cc.cryptopunks.astral/cc.cryptopunks.astral.service.ui.MainActivity
