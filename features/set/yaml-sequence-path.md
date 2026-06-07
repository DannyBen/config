# set/yaml-sequence-path

Set creates missing parents through YAML sequence indexes.

## Source Files

### YAML

```yaml
- name: admin
- name: user
```

## Commands

```shell
config set 0.pass hello
config set 1.auth.pass secret
config set auth.type basic --on name:user
```

## Result Files

### YAML

```yaml
- name: admin
  pass: hello
- name: user
  auth:
    pass: secret
    type: basic
```
