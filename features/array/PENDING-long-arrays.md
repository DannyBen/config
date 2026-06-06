# array/long-arrays

> PENDING Decide and implement the YAML long-array formatting threshold.

Replacing an entire array may reformat that array. Short scalar arrays should
use flow style. Longer scalar arrays should use block style for readability.

## Source Files

```yaml
title: yaml
users: [guest, admin]
```

```toml
title = "toml"
users = ["guest", "admin"]
```

## Commands

```shell
config array set users guest root auditor manager operator reviewer
```

## Result Files

```yaml
title: yaml
users:
  - guest
  - root
  - auditor
  - manager
  - operator
  - reviewer
```

```toml
title = "toml"
users = ["guest", "root", "auditor", "manager", "operator", "reviewer"]
```
