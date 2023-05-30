# filtering protocol (draft)
protocol for filtering data streams

## Data types

### matcher

|type|name|desc|
|-|-|-|
|[c]c|key|key used to obtain the metadata parameter value|
|[c]c|pattern|regex pattern for matching meta-data value|

### filter

|type|name|desc|
|-|-|-|
|\[c\][matcher](util-filter.md#matcher)|matchers|list of stream regex matchers|

