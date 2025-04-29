# objects

The objects module provides APIs for high-level operations on data objects.

## APIs

### objects.search

Search for objects matching a query.

Params:

| name   | descrption                                                            |
|:-------|:----------------------------------------------------------------------|
| q      | query string                                                          |
| zones  | optional: zones to include in the search                              |
| format | optional: alternative response format (only json is supported)        |
| ext    | optional: a comma-separated list of identities to consider as sources |

Example:

`objects.search?q=annual+report&zone=dvn&format=json&ext=alias1,key1`

Response is a stream of [SearchResult](search_result.go) objects ended with an EOF.
