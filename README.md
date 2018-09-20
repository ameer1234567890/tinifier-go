[![Build Status](https://travis-ci.org/ameer1234567890/tinifier-go.svg?branch=master)](https://travis-ci.org/ameer1234567890/tinifier-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ameer1234567890/tinifier-go)](https://goreportcard.com/report/github.com/ameer1234567890/tinifier-go)

#### Setup Instructions
* `git clone https://github.com/ameer1234567890/tinifier-go`
* `cd tinifier-go`
* If you are on Linux, run `go build -o $GOPATH/bin/tinifier tinifier.go`
* If you are on Windows, run `go build -o %GOPATH%/bin/tinifier.exe tinifier.go`
* Create a file named `.tinify_api_key` one level above the working directory. This file should contain the tinify.com API key without a leading CR and/or LF.
* Create two folders named `files` and `compressed` in the working folder.
* Place all pictures which needs to be compressed, inside the `files` folder.
* Run `tinifier`.
