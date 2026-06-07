# set/basic

Basic set operations

## Source Files

### YAML

```yaml
title: config
style:
  color: red
  size: 14

sections:
- caption: Getting Started
  tags: &basic [beginners, tutorial]
- caption: Installation
  tags: *basic
- caption: API
  tags: [dev, api]
```

### TOML

```toml
title = "config"
[style]
color = "red"
size = 14

[[sections]]
caption = "Getting Started"
tags = ["beginners", "tutorial"]

[[sections]]
caption = "Installation"
tags = ["beginners", "tutorial"]

[[sections]]
caption = "API"
tags = ["dev", "api"]
```

### JSON

```json
{
  "title": "config",
  "style": {
    "color": "red",
    "size": 14
  },
  "sections": [
    {
      "caption": "Getting Started",
      "tags": ["beginners", "tutorial"]
    },
    {
      "caption": "Installation",
      "tags": ["beginners", "tutorial"]
    },
    {
      "caption": "API",
      "tags": ["dev", "api"]
    }
  ]
}
```

## Commands

```shell
config list
-> title=config
-> style.color=red
-> style.size=14
-> sections.0.caption=Getting Started
-> sections.0.tags.0=beginners
-> sections.0.tags.1=tutorial
-> sections.1.caption=Installation
-> sections.1.tags.0=beginners
-> sections.1.tags.1=tutorial
-> sections.2.caption=API
-> sections.2.tags.0=dev
-> sections.2.tags.1=api

config list style
-> style.color=red
-> style.size=14

config list style.color
-> style.color=red
```
