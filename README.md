# dokku limit (beta)
App resource management plugin for dokku.

Currently supports settings limit's for memory usage and CPU usage.

## requirements

- dokku 0.4.x+
- docker 1.8.x

## installation

```shell
# on 0.4.x+
sudo dokku plugin:install https://github.com/sarendsen/dokku-limit.git limit
```

## Commands

```
limit (<app> <proc>), Pretty-print resource limits
limit:set <app> <proc> [memory=VALUE cpu=VALUE] [--no-restart], Set one or more limits for given app/process pair
```


## Resources

memory
cpu


## Usage

```
dokku limit:set my_app web cpu=50 memory=500m
```