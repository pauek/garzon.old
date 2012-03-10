#!/bin/sh

echo "Performing simple regression tests"
OS=`uname`
if [ "$OS" = "Linux" ]; then
  if [ `uname -m` = "x86_64" ]; then
    OS="Linux64"
  fi
fi
RES=0

for POL in *.policy.$OS; do
	PROG=`echo $POL | cut -f1 -d.`
	ARGS=""
	if [ -f $PROG.args ]; then
		ARGS=`cat $PROG.args`
	fi

	echo -n "$PROG:"
	SYSTR_RES=`eval ../systrace -f $POL -a $PROG $ARGS 2>/dev/null`
	# echo -e "\t(../systrace -f $POL -a $PROG $ARGS)"
	NORM_RES=`$PROG $ARGS`

	if [ -z "$SYSTR_RES" ] ; then
		rm -f id.core
		echo -e "\tFAILED"
		RES=1
	elif [ "$NORM_RES" != "$SYSTR_RES" ] ; then
		echo -e "\tExpected \"$NORM_RES\", got \"$SYSTR_RES\""
		RES=1
	else 
		echo -e "\tOKAY"
	fi

done

exit $RES
