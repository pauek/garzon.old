#!/bin/sh
path=$(echo $GOPATH | cut -d: -f1)
echo -n "grz-jail";   gcc -Wall -static -o $path/bin/grz-jail grz-eval/grz-jail/grz-jail.c
echo -n " grz";       go install ./grz
echo -n " grz-db";    go install ./grz-db
echo -n " grz-eval";  go install ./grz-eval
echo -n " grz-judge"; go install ./grz-judge
echo
