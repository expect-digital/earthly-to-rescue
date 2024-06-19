# Introduction

The purpose of this repo is to compare common automation tasks using [dagger](https://dager.io) and [earthly](https://earthly.dev) for a simple counter service that stores counter in [redis server](https://redis.io).

This repo contains [dagger](https://dagger.io) implementation.

The [earthly](https://earthly.dev) implementation is in [this repo](https://github.com/expect-digital/earthly-to-rescue).

# Build targets

# Outstanding tasks

## Function "image"

Saving image in local docker instance is not possible. Image is always published to container registry.

```shell
✗ dagger call image --src=.
✘ DaggerToRescue.image(
    src: ✔ ModuleSource.resolveDirectoryFromCaller(path: "."): Directory! 0.0s
  ): String! 1.8s
! call function "Image": process "/runtime" did not complete successfully: exit code: 2
┃ invoke: input: container.from.withFile.withEntrypoint.publish resolve: failed to export: failed to push counter/counter:latest: push access denied, repository does not exist or may require authorization: server message: insufficient_scope: authorization failed
┃
  ✘ Container.publish(address: "counter/counter:latest"): String! 0.9s
  ! failed to export: failed to push counter/counter:latest: push access denied, repository does not exist or may require authorization: server message: insufficient_scope: authorization failed
```

## Function "Up" and "Down"

Not possible? Dagger cannot execute on host environment, only in containers.
