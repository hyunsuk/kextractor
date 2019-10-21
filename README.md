# Kpick

[![Go Report Card](https://goreportcard.com/badge/github.com/loganstone/kpick)](https://goreportcard.com/report/github.com/loganstone/kpick)

날려 만들었으나 점점 개선하고 있는 한글 추출기

## Getting started

### Prerequisites

* [Install golang](https://golang.org/doc/install) ;)

### Install and run

* Install

```shell
$ go get github.com/loganstone/kpick
```

* Run

```shell
$ kpick -d /some-directory -f js

```

### Usage of kpick

```shell
$ kpick -h
Usage of kpick:
  -cpuprofile file
        Write cpu profile to file.
  -d string
        Directory to search. (default ".")
  -e    Make output error only.
  -f string
        File extension to scan. (default "*")
  -i    Interactive scanning.
  -igg string
        Pattern for line to ignore when scanning file.
  -memprofile file
        Write memory profile to file.
  -s string
        Directories to skip from search.(delimiter ',') (default ".git,tmp")
  -v    Make some output more verbose.

```

## Profiling example

```shell
$ kpick -d ../exchange-demo -f js -cpuprofile cpu.prof -memprofile mem.prof -igg //
$ go tool pprof -http 0.0.0.0:9000 cpu.prof
$ go tool pprof -http 0.0.0.0:9000 mem.prof
```

## Key Features

- 지정한 디렉터리에서 지정한 소스 파일들을 찾아 소스 파일 안에 한글이 있는
  줄과 줄 번호를 파일 이름이 포함된 경로와 함께 출력합니다.

## To-do Features

- Add something if need.
