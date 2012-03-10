#include <sys/types.h>

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif /* HAVE_CONFIG_H */

#include <stdio.h>
#include <sys/stat.h>
#include <sys/wait.h>
#include <unistd.h>
#include <assert.h>
#include <errno.h>
#include <stdlib.h>

#ifdef HAVE_SCHED_H
#include <sched.h>
#endif

void Test0(void)
{
	pid_t wpid;
	int status = -1;

	printf("\n-----Entering %s\n", __func__);
	
	printf("I am: %d\n", getpid());

	wpid = waitpid(-1, &status, 0);

	printf("Got wpid %d status %d\n", wpid, status);

	assert(wpid == -1);
	assert(errno == ECHILD);
}

void Test1(void)
{
	pid_t pid, wpid;
	int status = -1;

	printf("\n-----Entering %s\n", __func__);
	
	if ((pid = fork()) == 0) {
		printf("I am the child\n");
		exit(0);
	}

	printf("I am: %d\n", getpid());
	printf("Got child: %d\n", pid);

	wpid = waitpid(-1, &status, 0);

	printf("Got wpid %d status %d, wanted %d\n", wpid, status, pid);

	assert(wpid == pid);
	assert(status == 0);
}

void Test2(void)
{
	pid_t pid, wpid;
	int status = -1;

	printf("\n-----Entering %s\n", __func__);
	
	if ((pid = fork()) == 0) {
		printf("I am the child\n");
		sleep(1);
		exit(0);
	}

	printf("I am: %d\n", getpid());
	printf("Got child: %d\n", pid);

	while ((wpid = waitpid(-1, &status, WNOHANG)) == 0)
		;

	printf("Got wpid %d status %d, wanted %d\n", wpid, status, pid);

	assert(wpid == pid);
	assert(status == 0);
}

void Test3(void)
{
	pid_t pid, wpid;
	int status = -1;

	printf("\n-----Entering %s\n", __func__);

	setsid();
	setpgid(0, 0);
	
	if ((pid = fork()) == 0) {
		if (setpgid(0, 0) == -1)
			err(1, "%s: setpgid", __func__);
		printf("I am child one\n");
		exit(0);
	}

	if ((pid = fork()) == 0) {
		if (setpgid(0, getppid()) == -1)
			err(1, "%s: setpgid", __func__);
		printf("I am child two\n");
		sleep(1);
		exit(1);
	}
	
	printf("I am: %d\n", getpid());
	printf("Got child: %d\n", pid);

	while ((wpid = waitpid(0, &status, WNOHANG)) == 0)
		;

	printf("Got wpid %d status %d, wanted pid %d\n", wpid, status, pid);

	assert(wpid == pid);
	assert(WEXITSTATUS(status) == 1);

	wpid = waitpid(-1, &status, 0);
	printf("Got wpid %d status %d\n", wpid, status);
	assert(wpid > 0);
	assert(WEXITSTATUS(status) == 0);
}

#ifdef CLONE_THREAD
int childExecution(void *arg)
{
	printf("I am a thread: pid %d\n", getpid());
	
	sleep(1);

	if (arg == NULL) {
		sleep(10);
		return (1);
	}

	printf("I am exiting all of us.\n");
	exit(1);
}

#define STACK_SIZE 4096
char master_stack[STACK_SIZE];
char stacks[3][STACK_SIZE];

int masterExecution(void *arg)
{
	pid_t pid[3];
	int i;
	printf("I am a master thread: pid %d\n", getpid());
	
	for (i = 0; i < 3; ++i) {
		pid[i] = clone(childExecution, stacks[i] + STACK_SIZE,
		    SIGCHLD | CLONE_THREAD | CLONE_SIGHAND | CLONE_VM,
		    i == 0 ? (void *)0xdeadbeef : NULL);
		if (pid[i] == -1)
			err(1, "clone failed");
	}
	
	return (0);
}

void Test4(void)
{
	pid_t pid, wpid;
	int status;
	int i;
	
	printf("\n-----Entering %s\n", __func__);

	printf("I am: %d\n", getpid());

	pid = clone(masterExecution, master_stack + STACK_SIZE,
	    SIGCHLD | CLONE_SIGHAND | CLONE_VM, NULL);
	if (pid == -1)
		err(1, "clone failed");

	wpid = waitpid(-1, &status, __WCLONE | __WALL);
	if (wpid == -1)
		err(1, "waitpid");
	printf("Got wpid %d status %d, wanted %d\n", wpid, status, pid);

	wpid = waitpid(-1, &status, __WCLONE | __WALL);
	assert(wpid == -1);

	printf("Exiting.\n");
}
#endif

int
main(void)
{
	Test0();
	Test1();
	Test2();
	Test3();
#ifdef CLONE_THREAD
	Test4();
#endif       
        exit(0);
}
