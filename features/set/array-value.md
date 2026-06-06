# set/array-value

Insert and update arrays

## Source Files

```yaml
title: yaml
ports: [3000, 4000]
methods: [GET, SET]
users:
  - guest
  - admin
```

```toml
title = "toml"
ports = [3000, 4000]
methods = ["GET", "SET"]
users = ["guest", "admin"]
```

## Commands

```shell
config set ports 80 8080
config set users guest root
config set tags api worker
config set methods.1 POST
config set aliases PUT --array
```

## Result Files

```yaml
title: yaml
ports: [80, 8080]
methods: [GET, POST]
users: [guest, root]
tags: [api, worker]
aliases: [PUT]
```

```toml
title = "toml"
ports = [80, 8080]
methods = ["GET", "POST"]
users = ["guest", "root"]
tags = ["api", "worker"]
aliases = ["PUT"]
```
