#!/bin/sh
# read-remote-object has no run-phase setup: node2 has no Qwen operator, so the
# remote read of node1's object IS the thing under test. verify.py performs it
# (resolve node1's identity + the object_id stored by object-store, then have node2
# read <node1>:objects.load and assert the bytes). run.sh is a no-op placeholder so
# netsim discovers the task and hands off to verify.sh.
set -eu
echo "read-remote-object: no run-phase setup; verify.py performs the cross-read."
