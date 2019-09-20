# Description

This repository intended to test can BadgerDB read wrong/incomlete data without any errors.

## How to run

    ./all-versions.sh -count=100 -failfast

## Possible issue

One case that is probably real:

* We have directory with valid Badger data files: sst files, vlog file.
* We change bit of value (spoil data) in vlog file.
* Badger starts with provided vlog file and end user got database with spoiled value, without any notice.

## Gory details

* make.go builds datagen script for required badger version (supports 1.5, 1.6, 2.0)
* make.go run test in _spoiler_tests
* TODO: _spoiler_tests description
