# array/long-arrays

YAML array style follows the final array shape and the existing source style.
New arrays are written in block style. Existing block arrays stay block style.
Existing flow arrays stay flow style until they grow to five values.

## Source Files

### YAML

```yaml
title: yaml
tags: [dev, stage]
ports:
  - 3000
  - 4000
```

### TOML

```toml
title = "toml"
tags = ["dev", "stage"]
ports = [3000, 4000]
```

### JSON

```json
{
  "title": "json",
  "tags": ["dev", "stage"],
  "ports": [3000, 4000]
}
```

## Commands

```shell
config array add tags api sales db
config array add ports 5000
```

## Result Files

### YAML

```yaml
title: yaml
tags:
  - dev
  - stage
  - api
  - sales
  - db
ports:
  - 3000
  - 4000
  - 5000
```

### TOML

```toml
title = "toml"
tags = ["dev", "stage", "api", "sales", "db"]
ports = [3000, 4000, 5000]
```

### JSON

```json
{
  "ports": [
    3000,
    4000,
    5000
  ],
  "tags": [
    "dev",
    "stage",
    "api",
    "sales",
    "db"
  ],
  "title": "json"
}
```
