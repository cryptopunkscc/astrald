# install-astrald

Builds `astrald` and `astral-query` from source and runs `astrald` as a systemd
service on the target VMs (all running VMs by default, or `--vm <host>`; `--ref`
builds a specific git ref). `verify.sh` confirms each node answers
`astral-query localnode:.spec`. The service is left running so the netsim stage
snapshots a live node that resumes already-running. Used by `lab.story`; see
[Running astrald as a service](../../../docs/running-as-a-service.md) for the unit
file and operational details.
