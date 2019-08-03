[![CircleCI](https://circleci.com/gh/mlabouardy/nexus-cli.svg?style=svg)](https://circleci.com/gh/mlabouardy/nexus-cli) [![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

<div align="center">
<img src="logo.png" width="60%"/>
</div>

Nexus CLI for Docker Registry

## Usage

<div align="center">
<img src="example.png"/>
</div>

## Installation

To install the library and command line program, use the following:

```bash
$ go get -u github.com/f9n/nexus-cli
```

## Available Commands

```bash
$ nexus-cli configure
```

```bash
$ nexus-cli image ls
$ nexus-cli image ls --detail
$ nexus-cli image ls --sort-by-size
```

```bash
$ nexus-cli image tags -name mlabouardy/nginx
```

```bash
$ nexus-cli image info -name mlabouardy/nginx -tag 1.2.0
```

```bash
$ nexus-cli image delete -name mlabouardy/nginx
$ nexus-cli image delete -name mlabouardy/nginx -tag 1.2.0
$ nexus-cli image delete -name mlabouardy/nginx -keep 4
```

```bash
$ nexus-cli image size -name mlabouardy/nginx
$ nexus-cli image size -name mlabouardy/nginx --human-readable
```
## Tutorials

* [Cleanup old Docker images from Nexus Repository](http://www.blog.labouardy.com/cleanup-old-docker-images-from-nexus-repository/)
