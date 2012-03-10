#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>

int
main(void)
{
        struct stat sb;
        int err;

        if (stat("", &sb) == -1)
		printf("OKAY\n");
	else
		printf("FAILED\n");

        exit(0);
}
