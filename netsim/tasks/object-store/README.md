# object-store

The operator on node1 stores `payload.txt` as an astral object on `--target` (default `localnode`; an alias like `node2` stores on that peer) and records the id. `verify.py` loads the object from the holder's local repo and asserts the bytes equal `payload.txt`.

Stories: `object-store.story` (`localnode`) produces `two-nodes-data` and feeds `read-remote-object`; `object-store-peer.story` (`--target node2`) produces `two-nodes-data-peer`.
