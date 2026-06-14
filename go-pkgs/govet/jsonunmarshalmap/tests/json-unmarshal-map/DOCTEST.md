# JSON Unmarshal to Map Detection Tests

Tests for the `json-unmarshal-map` checker that detects `json.Unmarshal` calls
targeting `map[string]any` or `map[string]interface{}` instead of typed structs.

## How to Run

```sh
doctest test -v ./
```
