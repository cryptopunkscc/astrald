# install-astrald

Builds `astrald` and `astral-query` from source and runs `astrald` as a systemd service on the target VMs (all running, or `--vm <host>`; `--ref` picks a git ref). Verify proves the unit is enabled and each node answers `astral-query localnode:.spec`. Left running for the snapshot; used by `lab.story`.
