#!/bin/sh

path=$(echo $GOPATH | cut -d: -f1)
gcc -Wall -static -o $path/bin/grz-jail grz-jail/grz-jail.c
go install ./grz
go install ./grz-eval
go install ./grz-judge
