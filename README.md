# kpick
날려 만들었으나 점점 개선하고 있는 한글 pick

## Getting started

### Install

* Install golang ;)

* Install `kpick`

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

## Key Features

- 지정한 디렉터리에서 지정한 소스 파일을 찾아
  소스 파일 안에 한글이 있는 부분을 알려줍니다.

## To-do Features

- Add something if need.
