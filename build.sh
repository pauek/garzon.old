#!/bin/sh
echo -n "grz-jail";   go install ./grz-jail
echo -n " grz";       go install ./grz
echo -n " grz-db";    go install ./grz-db
echo -n " grz-eval";  go install ./grz-eval
echo    " grz-judge"; go install ./grz-judge
