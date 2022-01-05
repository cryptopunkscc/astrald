# Warp Drive CLI

Command line service for warpdrive, combined with anc can be considered as referential UI client. Translates text
commands into warpdrive UI API calls, and formats responses into console output.

### Connect console session

```shell
$ anc query <identity> wd
connected.
```

## Features

## `sender`

### `recipients`

```shell
warp> peers
02f978d6bd70d0005f5148ce5h311609c994219164126488f11573440b2c6a40eb localhost  
024d47047667312be7cd0a140f3323b716030f5fc9d37ae774eb96527a76fa55f9 remote
ok
```

### `send`

```shell
warp> send <path/to/file/or/directory> <peer_id>
767a74a1-afb4-4a5d-6f0c-0ca8d3f25a1d
ok
```

### `sent`

```shell
warp> sent
3a076839-1b17-41a9-50b9-72beee2d08db  sent
  -  wdtest
  -  wdtest/foo
  -  wdtest/test
  -  wdtest/test/bar
767a74a1-afb4-4a5d-6f0c-0ca8d3f25a1d  uploaded
  -  baz
767a74a1-afb4-4a5d-6f0c-0ca8d3f25a1d  rejected
  -  fiz
ok
```

### `events`

```shell
warp> events sent
f07a618b-a45d-459f-7d05-98a329bbe3e9 sent
f07a618b-a45d-459f-7d05-98a329bbe3e9 rejected
34ca2937-daef-4949-59f0-625170569eac sent
34ca2937-daef-4949-59f0-625170569eac accepted
34ca2937-daef-4949-59f0-625170569eac downloaded
```

## `recipient`

### `received`

```shell
warp> sent
3a076839-1b17-41a9-50b9-72beee2d08db  sent
  -  wdtest
  -  wdtest/foo
  -  wdtest/test
  -  wdtest/test/bar
```

### `accept`

```shell
warp> accept f24d4a5f-2d84-4c15-5382-f21600016f4c
accepted f24d4a5f-2d84-4c15-5382-f21600016f4c
ok
```

### `reject`

```shell
warp> reject 311b3b74-3bfa-44de-75b3-5ffdcab742fb
rejected 311b3b74-3bfa-44de-75b3-5ffdcab742fb
ok
```

### `update`

```shell
warp> update 024d47047667312ba6cd0a14033423b716030f5fc9dd7ae774eb96527a76fa55f9 mod trust
updated [024d47047667312ba6cd0a14033423b716030f5fc9dd7ae774eb96527a76fa55f9 mod trust]
ok
```

```shell
warp> update 024d47047667312ba6cd0a14033423b716030f5fc9dd7ae774eb96527a76fa55f9 mod block
updated [024d47047667312ba6cd0a14033423b716030f5fc9dd7ae774eb96527a76fa55f9 mod block]
ok
```

### `events`

```shell
warp> events received
f07a618b-a45d-459f-7d05-98a329bbe3e9 received
f07a618b-a45d-459f-7d05-98a329bbe3e9 rejected
c7137926-3c90-4629-4179-3bcb0ca9bf6d received
c7137926-3c90-4629-4179-3bcb0ca9bf6d accepted
c7137926-3c90-4629-4179-3bcb0ca9bf6d downloaded
```