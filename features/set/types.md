# set/types

Set preserves types

## Source Files

```yaml
title: yaml
```

```toml
title = "toml"
```

## Commands

```shell
config set bool true
config set date 2024-01-13
config set time 14:30:00
config set timestamp 2027-03-24T14:30:00Z
config set int 10
config set negative -- -10
config set float 1.0
config set exponent 1e6
config set string-fallback localhost
config set string-number 1.0 --string
config set string-date 2024-01-13 --string
```

## Result Files

```yaml
title: yaml
bool: true
date: 2024-01-13
time: 14:30:00
timestamp: 2027-03-24T14:30:00Z
int: 10
negative: -10
float: 1.0
exponent: 1e6
string-fallback: localhost
string-number: "1.0"
string-date: "2024-01-13"
```

```toml
title = "toml"
bool = true
date = 2024-01-13
time = 14:30:00
timestamp = 2027-03-24T14:30:00Z
int = 10
negative = -10
float = 1.0
exponent = 1e6
string-fallback = "localhost"
string-number = "1.0"
string-date = "2024-01-13"
```
