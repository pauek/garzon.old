package main

/*

#include <termios.h>
#include <unistd.h>

struct termios termios_p;

void toggle_echo() {
   tcgetattr(0, &termios_p);
   termios_p.c_lflag ^= ECHO;
   tcsetattr(0, TCSANOW, &termios_p);
}

*/
import "C"

func ToggleEcho() {
	C.toggle_echo()
}