#!/bin/bash

test -f ./build/astral-android.apk || ./buildApk.sh


if [ $1 == -r ]; then
  adb uninstall cc.cryptopunks.astral
fi

adb install build/astral-android.apk
adb shell am start -n cc.cryptopunks.astral/cc.cryptopunks.astral.ui.main.MainActivity
