#!/bin/sh
echo -n "grz-jail";   go install ./grz-jail
echo -n " grz";       go install ./grz
echo -n " grz-db";    go install ./grz-db
echo -n " grz-eval";  go install ./grz-eval
echo -n " grz-judge"; go install ./grz-judge
echo
