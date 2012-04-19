package main

/*

#include <termios.h>
#include <unistd.h>

struct termios termios_p;

void disable_echo() {
   tcgetattr(0, &termios_p);
   termios_p.c_lflag &= ~ECHO;
   tcsetattr(0, TCSANOW, &termios_p);
}

void enable_echo() {
   tcgetattr(0, &termios_p);
   termios_p.c_lflag |= ECHO;
   tcsetattr(0, TCSANOW, &termios_p);
}

*/
import "C"

func EnableEcho() {
	C.enable_echo()
}

func DisableEcho() {
	C.disable_echo()
}
