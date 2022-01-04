#!/bin/bash

test -f ./libs/astral-android.apk || ./buildApk.sh

adb install libs/astral-android.apk
adb shell am start -n cc.cryptopunks.astral.node/cc.cryptopunks.astral.service.ui.MainActivity
