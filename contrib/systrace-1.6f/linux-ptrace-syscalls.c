/*
 * Copyright (c) 2002 Marius Aamodt Eriksen <marius@umich.edu>
 * Copyright (c) 2002-2006 Niels Provos <provos@citi.umich.edu>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software
 *    must display the following acknowledgement:
 *      This product includes software developed by Marius Aamodt Eriksen
 *      and Niels Provos.
 * 4. The name of the author may not be used to endorse or promote products
 *    derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
 * IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
 * OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 * IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
 * INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
 * NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
 * THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */


#include <sys/types.h>

typedef u_int32_t u32;

#include <asm/unistd.h>

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif /* HAVE_CONFIG_H */

#include <sys/socket.h>
#include <sys/ioctl.h>
#include <sys/ptrace.h>
#include <sys/wait.h>

#include <linux/limits.h>
#include <linux/types.h>
#ifdef HAVE_LINUX_USER_H
#include <linux/user.h>
#else
#ifdef HAVE_SYS_USER_H
#include <sys/user.h>
#else
#ifdef HAVE_ASM_USER_H
#include <asm/user.h>
#endif  /* HAVE_ASM_USER_H */
#endif  /* HAVE_SYS_USER_H */
#endif  /* HAVE_LINUX_USER_H */
#include <linux/ptrace.h>	/* for PTRACE_O_TRACESYSGOOD */
#include <sys/queue.h>
#include <sys/tree.h>

#define MAX_SYSCALLS 2048	/* maximum number of system calls we support */

#include <limits.h>
#include <err.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#ifdef HAVE_SCHED_H
#include <sched.h>
#else
#ifdef    HAVE_LINUX_SCHED_H
#include <linux/sched.h>
#endif    /* HAVE_LINUX_SCHED_H */
#endif

#include <event.h>
#include <errno.h>

#undef DEBUG
#ifdef DEBUG
#define DFPRINTF(x)	fprintf x
#else
#define DFPRINTF(x)
#endif

#include "intercept.h"
#include "linux_syscalls.c"
#include "systrace-errno.h"

#ifdef PTRACE_LINUX64
#define ISLINUX32(x)		(linux_call_type((x)->cs) == LINUX32)
#define SYSCALL_NUM(x)		(x)->orig_rax
#define SET_RETURN_CODE(x, v)	(x)->rax = (v)
#define RETURN_CODE(x)		(ISLINUX32(x) ? (long)(int)(x)->rax : (x)->rax)
#define ARGUMENT_0(x)		(ISLINUX32(x) ? (x)->rbx : (x)->rdi)
#define ARGUMENT_1(x)		(ISLINUX32(x) ? (x)->rcx : (x)->rsi)
#define ARGUMENT_2(x)		(ISLINUX32(x) ? (x)->rdx : (x)->rdx)
#define ARGUMENT_3(x)		(ISLINUX32(x) ? (x)->rsi : (x)->rcx)
#define ARGUMENT_4(x)		(ISLINUX32(x) ? (x)->rdi : (x)->r8)
#define ARGUMENT_5(x)		(ISLINUX32(x) ? (x)->rbp : (x)->r9)
#define SET_ARGUMENT_0(x, v)	if (ISLINUX32(x)) (x)->rbx = (v); else (x)->rdi = (v)
#define SET_ARGUMENT_1(x, v)	if (ISLINUX32(x)) (x)->rcx = (v); else (x)->rsi = (v)
#define SET_ARGUMENT_2(x, v)	if (ISLINUX32(x)) (x)->rdx = (v); else (x)->rdx = (v)
#define SET_ARGUMENT_3(x, v)	if (ISLINUX32(x)) (x)->rsi = (v); else (x)->rcx = (v)
#define SET_ARGUMENT_4(x, v)	if (ISLINUX32(x)) (x)->rdi = (v); else (x)->r8 = (v)
#define SET_ARGUMENT_5(x, v)	if (ISLINUX32(x)) (x)->rbp = (v); else (x)->r9 = (v)
#else
#define SYSCALL_NUM(x)		(x)->orig_eax
#define SET_RETURN_CODE(x, v)	(x)->eax = (v)
#define RETURN_CODE(x)		(x)->eax
#define ARGUMENT_0(x)		(x)->ebx
#define ARGUMENT_1(x)		(x)->ecx
#define ARGUMENT_2(x)		(x)->edx
#define ARGUMENT_3(x)		(x)->esi
#define ARGUMENT_4(x)		(x)->edi
#define ARGUMENT_5(x)		(x)->ebp
#define SET_ARGUMENT_0(x, v)	(x)->ebx = (v)
#define SET_ARGUMENT_1(x, v)	(x)->ecx = (v)
#define SET_ARGUMENT_2(x, v)	(x)->edx = (v)
#define SET_ARGUMENT_3(x, v)	(x)->esi = (v)
#define SET_ARGUMENT_4(x, v)	(x)->edi = (v)
#define SET_ARGUMENT_5(x, v)	(x)->ebp = (v)
#endif /* !PTRACE_LINUX64 */

static int                   linux_init(void);
static int                   linux_attach(int, pid_t);
static int                   linux_report(int, pid_t);
static int                   linux_detach(int, pid_t);
static int                   linux_open(void);
static struct intercept_pid *linux_getpid(pid_t);
static void                  linux_freepid(struct intercept_pid *);
static void                  linux_clonepid(struct intercept_pid *, struct intercept_pid *);
static const char           *linux_syscall_name(enum LINUX_CALL_TYPES, pid_t, int);
static int                   linux_syscall_number(const char *, const char *);
static short                 linux_translate_flags(short);
static int                   linux_translate_errno(int);
static int                   linux_answer(int, pid_t, u_int32_t, short, int, short, struct elevate *);
static int                   linux_newpolicy(int);
static int                   linux_assignpolicy(int, pid_t, int);
static int                   linux_modifypolicy(int, int, int, short);
static int                   linux_replace(int, pid_t, u_int16_t, struct intercept_replace *);
static int                   linux_io(int, pid_t, int, void *, u_char *, size_t);
static int                   linux_setcwd(int, pid_t);
static int                   linux_restcwd(int);
static int                   linux_argument(int, void *, int, void **);
static int                   linux_read(int);
static int                   linux_isfork(const char *);

static void		     linux_remove_pidstatus(struct intercept_pid *, pid_t);
static void		     linux_wakeprocesses(struct intercept_pid *, pid_t cpid);

static void                  linux_atexit(void);

/* Local state */
#define SYSTR_POLICY_ASK	0
#define SYSTR_POLICY_PERMIT	1
#define SYSTR_POLICY_NEVER	2

#define SYSTR_FLAGS_RESULT		0x001
#define SYSTR_FLAGS_SAWEXECVE		0x002
#define SYSTR_FLAGS_SAWFORK		0x004
#define SYSTR_FLAGS_SKIPSTOP		0x008
#define SYSTR_FLAGS_SAWWAITPID		0x010
#define SYSTR_FLAGS_ERRORCODE		0x020
#define SYSTR_FLAGS_PAUSING		0x040
#define SYSTR_FLAGS_STOPWAITING		0x080

#define SYSTR_FLAGS_CLONE_THREAD	0x100
#define SYSTR_FLAGS_CLONE_DETACHED	0x200
#define SYSTR_FLAGS_CLONE_EXITING	0x400

struct linux_policy {
	long error_code[MAX_SYSCALLS];
};

#define MAX_POLICIES	500
int policy_used[MAX_POLICIES];
static struct linux_policy policies[MAX_POLICIES];

struct linux_wait_pid {
	TAILQ_ENTRY(linux_wait_pid) next;

        int status;
        pid_t pid;
        pid_t pgid;
};

struct linux_data {
	enum { SYSCALL_START = 0, SYSCALL_END } status;
	int policy;
	int flags;
	long error_code;
	long pstatus;	/* where to stick the return code */
	int wstatus;	/* saved error code for exit */
	pid_t waitpid;
	pid_t pgid;	/* process group tracking */

	struct user_regs_struct regs;

	/* everything below here needs to be cleared on clone */
        int nchildren;
        int nthreads;
	int nthreads_waiting;
        int nthreads_detached;

	TAILQ_HEAD(childq, linux_wait_pid) waitq;
};

static int notify_fd = -1; 

enum LINUX_SYSCALLS {
	SYSTR_CLONE = 0,
	SYSTR_GETPID = 1,
	SYSTR_WAIT4 = 2,
	SYSTR_EXECVE = 3,
	SYSTR_SETSID = 4,
	SYSTR_SETPGID = 5,
	SYSTR_NUM_CALLS = 6
};

/* mapps special system calls used internally by the ptrace backend to
 * the system call number. */
static int
linux_map_call(enum LINUX_CALL_TYPES call_type, enum LINUX_SYSCALLS call_number)
{
	static int initialized;
	static int mappings[LINUX_NUM_VERSIONS][SYSTR_NUM_CALLS];
	if (!initialized) {
		const char *names[] = {
			"clone", "getpid", "wait4", "execve", "setsid", "setpgid",
			NULL
		};
		
		int i;
		for (i = 0; names[i] != NULL; ++i) {
			int res = linux_syscall_number("linux", names[i]);
			assert(res != -1);
			mappings[LINUX32][i] = res;
			res = linux_syscall_number("linux64", names[i]);
			assert(res != -1);
			mappings[LINUX64][i] = res;
		}
		initialized = 1;
	}

	assert(call_type >= 0);
	assert(call_type < LINUX_NUM_VERSIONS);
	assert(call_number >= 0);
	assert(call_number < SYSTR_NUM_CALLS);

	return (mappings[call_type][call_number]);
}



static void
linux_term_signal(int fd, short what, void *arg)
{
	/* We intercepted a termination signal and should exit */
	exit(2);
}

static int
linux_init(void)
{
	static struct event sigterm_ev, sigint_ev;
	memset(policy_used, 0, sizeof(policy_used));

        /*
         * TODO: find a better way to do this.
         *
         * Using atexit is kind of ugly, since it basically requires
         * that systrace always dies in a way that gets the atexit
         * handlers to be called, but without kernel support it is
         * hard to make something entirely foolproof here.  It doesn't
         * seem that ptrace itself supports anything of the like.
         */
        if (atexit(&linux_atexit) < 0)
                errx(1, "atexit");

	signal_set(&sigint_ev, SIGINT, linux_term_signal, NULL);
	signal_add(&sigint_ev, NULL);
	signal_set(&sigterm_ev, SIGTERM, linux_term_signal, NULL);
	signal_add(&sigterm_ev, NULL);
	
	return (0);
}

/*
 * Ptrace is a pretty horrid interface for doing system call interposition.
 * We can differentiate between a SIGTRAP and the SIGTRAP we get when
 * intercepting a system call by setting the PTRACE_O_TRACESYSGOOD option.
 */

static int sigsyscall = SIGTRAP;

static int
linux_attach(int fd, pid_t pid)
{
	int status = 0;
	int res = ptrace(PTRACE_ATTACH, pid, NULL, NULL);
	if (res == -1) {
		warn("%s:%d %s: ptrace", __FILE__, __LINE__, __func__);
		return (-1);
	}

	res = waitpid(pid, &status, 0);
	if (res == -1 || !WIFSTOPPED(status)) {
		warn("%s:%d %s: waitpid: %d", __FILE__, __LINE__, __func__, status);
		ptrace(PTRACE_KILL, pid, NULL, NULL);
		return (-1);
	}

#ifdef PTRACE_O_TRACESYSGOOD
	DFPRINTF((stderr, "%s: setting TRACESYSGOOD\n", __func__));
	res = ptrace(PTRACE_SETOPTIONS, pid,
	    NULL, (void *)PTRACE_O_TRACESYSGOOD);
	if (res == -1) {
		warn("%s:%d %s: PTRACE_O_TRACESYSGOOD failed",
		    __FILE__, __LINE__, __func__);
	} else {
		sigsyscall |= 0x80;
	}
#endif
	
	/* continue the child so that it can do something */
	res = ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
	if (res == -1)
		err(1, "%s: ptrace", __func__);

	/* kick our notification loop */
	write(notify_fd, "C", 1);

	return (res);
}

static int
linux_report(int fd, pid_t pid)
{
	DFPRINTF((stderr, "%s: pid %d\n", __func__, pid));
	return (0);
}

static int
linux_detach(int fd, pid_t pid)
{
	DFPRINTF((stderr, "%s: pid %d\n", __func__, pid));
	return (ptrace(PTRACE_DETACH, pid, (char *)1, 0));
}

static int
linux_open(void)
{
	int pair[2];
	int fd = -1;

	if (notify_fd != -1)
		errx(1, "already initialized.");
	
	if (socketpair(AF_UNIX, SOCK_STREAM, 0, pair) == -1)
		err(1, "socketpair");

	fd = pair[0];
	notify_fd = pair[1];

	if (fcntl(fd, F_SETFD, 1) == -1)
		warn("%s:%d fcntl(F_SETFD)", __FILE__, __LINE__);
	if (fcntl(notify_fd, F_SETFD, 1) == -1)
		warn("%s:%d fcntl(F_SETFD)", __FILE__, __LINE__);

	return (fd);
}

static struct intercept_pid *
linux_getpid(pid_t pid)
{
	struct intercept_pid *icpid;
	struct linux_data *data;

	icpid = intercept_getpid(pid);
	if (icpid == NULL)
		return (NULL);
	if (icpid->data != NULL)
		return (icpid);

	if ((icpid->data = malloc(sizeof(struct linux_data))) == NULL)
		err(1, "%s:%d: malloc", __func__, __LINE__);

	data = icpid->data;
	memset(data, 0, sizeof(struct linux_data));

	/* no policy initially assigned */
	data->policy = -1;

	/* no waitpid available on start */
	data->waitpid = -1;

	/* no process group */
	data->pgid = -1;
	
	TAILQ_INIT(&data->waitq);
	
	return (icpid);
}

static void
linux_freepid(struct intercept_pid *ipid)
{
	struct linux_data* data = ipid->data;
	struct linux_wait_pid *wid;
	while ((wid = TAILQ_FIRST(&data->waitq)) != NULL) {
		TAILQ_REMOVE(&data->waitq, wid, next);
		free(wid);
	}
	if (ipid->data != NULL)
		free(ipid->data);
	ipid->data = NULL;
}

static void
linux_clonepid(struct intercept_pid *opid, struct intercept_pid *npid)
{
	struct linux_data *data = NULL;
	DFPRINTF((stderr, "%s: cloning for %d\n", __func__, opid->pid));
	if (opid->data == NULL) {
		npid->data = NULL;
		return;
	}

	if ((npid->data = malloc(sizeof(struct linux_data))) == NULL)
		err(1, "%s:%d: malloc", __func__, __LINE__);
	memcpy(npid->data, opid->data, sizeof(struct linux_data));

	/* We need to reset some internal states */
	data = npid->data;
	TAILQ_INIT(&data->waitq);

	data->nchildren = 0;
	data->nthreads = data->nthreads_detached = 0;
	data->nthreads_waiting = 0;
}

static struct linux_data *
linux_get_piddata(pid_t pid)
{
	struct intercept_pid *icpid = intercept_findpid(pid);
	return icpid != NULL ? icpid->data : NULL;
}

/*
 * returns potentially queued up status. if dofree is specified it fees
 * the entry.
 */

static pid_t
linux_find_pidstatus(struct intercept_pid *pid, pid_t cpid, int *status)
{
	struct linux_wait_pid *wid;
	struct linux_data *data = pid->data;

	DFPRINTF((stderr, "%s: pid %d (pgid %d) waits for %d\n",
		__func__, pid->pid, data->pgid, cpid));

	/* if this thread is a cloner, then we need to go to the parent */
	if (data->flags & SYSTR_FLAGS_CLONE_THREAD) {
		pid = intercept_findpid(pid->ppid);
		if (pid == NULL)
			errx(1, "%s: intercept_findpid", __func__);
		data = pid->data;
	}
	
	TAILQ_FOREACH(wid, &data->waitq, next) {
		/*
		 * For a cpid != -1, we really need to implement process
		 * group tracking.  So, do it.  0 means waiting for our
		 * own process group, a negative numbers means, wait for
		 * a specific process group.
		 */
		if (wid->pid == cpid ||
		    cpid == -1 ||
		    (cpid == 0 && wid->pgid == data->pgid) ||
		    (cpid < 0 && wid->pgid == -cpid)) {
			cpid = wid->pid;
			if (status != NULL)
				*status = wid->status;
			return (cpid);
		}
	}

	return (-1);
}

static void
linux_remove_pidstatus(struct intercept_pid *pid, pid_t cpid)
{
	struct linux_wait_pid *wid;
	struct linux_data *data = pid->data;

	/* if this thread is a cloner, then we need to go to the parent */
	if (data->flags & SYSTR_FLAGS_CLONE_THREAD) {
		pid = intercept_findpid(pid->ppid);
		if (pid == NULL)
			errx(1, "%s: intercept_findpid", __func__);
		data = pid->data;
	}

	DFPRINTF((stderr, "%s: pid %d (pgid %d) removes status for %d\n",
		__func__, pid->pid, data->pgid, cpid));
	TAILQ_FOREACH(wid, &data->waitq, next) {
		if (wid->pid == cpid) {
			TAILQ_REMOVE(&data->waitq, wid, next);
			free(wid);
			return;
		}
	}

	errx(1, "%s: cannot find status for %d", __func__, cpid);
}

static void
linux_add_pidstatus(struct intercept_pid *pid,
    pid_t cpid, pid_t pgid, int status)
{
	struct linux_wait_pid *wid = NULL;
	struct linux_data *data = pid->data;

	DFPRINTF((stderr, "%s: pid %d get status of %d%s\n",
		__func__, pid->pid, cpid,
		data->flags & CLONE_DETACHED ? ": ignored" : ""));
	
	if (data->flags & CLONE_DETACHED)
		return;
	
	if ((wid = malloc(sizeof(*wid))) == NULL)
		err(1, "%s: malloc", __func__);
	
	wid->status = status;
	wid->pid = cpid;
	wid->pgid = pgid;

	TAILQ_INSERT_TAIL(&data->waitq, wid, next);

	/* Maybe the parent gets to wait now */
	linux_wakeprocesses(pid, cpid);
}

static void
linux_childdead(pid_t pid, int status)
{
	struct intercept_pid *icpid;
	struct linux_data *data;
	struct linux_data *parent_data;
	pid_t ppid;
	int reap_parent = 0;
	int already_detached = 0;
	
	if ((icpid = intercept_findpid(pid)) == NULL)
		err(1, "%s: intercept_getpid", __func__);
	data = icpid->data;

	DFPRINTF((stderr, "%s: pid %d (ppid %d, pgid %d): chld %d(%d)\n",
		__func__, pid, icpid->ppid, data->pgid,
		data->nchildren, data->nthreads));

	if (data->nthreads > 0) {
		DFPRINTF((stderr, "%s: pid %d waiting for %d children\n",
			__func__, pid, data->nthreads));
		/*
		 * we have more threads in this thread group and need
		 * to hang around until all are gone.
		 */
		data->flags |= SYSTR_FLAGS_CLONE_EXITING;
		data->wstatus = status;
		goto out;
	}

	if (data->flags & SYSTR_FLAGS_CLONE_EXITING)
		already_detached = 1;

	ppid = icpid->ppid;

	/* if there is no parent then things are easy */
	if (!ppid || (icpid = intercept_findpid(ppid)) == NULL) {
		DFPRINTF((stderr, "%s:  no parent to wait for %d\n",
			__func__, pid));
		intercept_child_info(pid, -1);
		goto out;
	}
	parent_data = icpid->data;
	
	/* one child less */
	parent_data->nchildren--;

	if (data->flags & SYSTR_FLAGS_CLONE_DETACHED)
		parent_data->nthreads_detached--;
	
	if (data->flags & SYSTR_FLAGS_CLONE_THREAD)
		parent_data->nthreads--;
	
	DFPRINTF((stderr, "%s: parent %d: remaining children %d, threads %d\n",
		__func__, ppid, parent_data->nchildren, parent_data->nthreads));
	
	linux_add_pidstatus(icpid, pid, data->pgid, status);

	/* let's see if our parent can go now, too */
	if ((data->flags & SYSTR_FLAGS_CLONE_THREAD)  &&
	    (parent_data->flags & SYSTR_FLAGS_CLONE_EXITING) &&
	    parent_data->nthreads == 0)
		reap_parent = 1;

	/* inform everyone about the child dying */
	intercept_child_info(pid, -1);

	if (reap_parent)
		linux_childdead(ppid, parent_data->wstatus);
	
out:
	if (!already_detached) {
		if (ptrace(PTRACE_DETACH, pid, (char *)1, 0) == -1) {
			if (errno != ESRCH)
				err(1, "%s: ptrace(DETACH)", __func__);
			if (kill(pid, 0) == -1) {
				if (errno != ESRCH)
					err(1, "%s: kill", __func__);
			}
		}
	}
}

static short
linux_translate_flags(short flags)
{
	switch (flags) {
	case ICFLAGS_RESULT:
		return (SYSTR_FLAGS_RESULT);
	default:
		return (0);
	}
}

static int
linux_translate_errno(int errnumber)
{
	/* XXX kind of nasty; make a table instead.  at least it was
	 * automatically generated. */

	switch (errnumber) {
	case SYSTRACE_EPERM: return (-1);
	case SYSTRACE_ENOENT: return (-2);
	case SYSTRACE_ESRCH: return (-3);
	case SYSTRACE_EINTR: return (-4);
	case SYSTRACE_EIO: return (-5);
	case SYSTRACE_ENXIO: return (-6);
	case SYSTRACE_E2BIG: return (-7);
	case SYSTRACE_ENOEXEC: return (-8);
	case SYSTRACE_EBADF: return (-9);
	case SYSTRACE_ECHILD: return (-10);
	case SYSTRACE_EAGAIN: return (-11);
	case SYSTRACE_ENOMEM: return (-12);
	case SYSTRACE_EACCES: return (-13);
	case SYSTRACE_EFAULT: return (-14);
	case SYSTRACE_ENOTBLK: return (-15);
	case SYSTRACE_EBUSY: return (-16);
	case SYSTRACE_EEXIST: return (-17);
	case SYSTRACE_EXDEV: return (-18);
	case SYSTRACE_ENODEV: return (-19);
	case SYSTRACE_ENOTDIR: return (-20);
	case SYSTRACE_EISDIR: return (-21);
	case SYSTRACE_EINVAL: return (-22);
	case SYSTRACE_ENFILE: return (-23);
	case SYSTRACE_EMFILE: return (-24);
	case SYSTRACE_ENOTTY: return (-25);
	case SYSTRACE_ETXTBSY: return (-26);
	case SYSTRACE_EFBIG: return (-27);
	case SYSTRACE_ENOSPC: return (-28);
	case SYSTRACE_ESPIPE: return (-29);
	case SYSTRACE_EROFS: return (-30);
	case SYSTRACE_EMLINK: return (-31);
	case SYSTRACE_EPIPE: return (-32);
	case SYSTRACE_EDOM: return (-33);
	case SYSTRACE_ERANGE: return (-34);
	case SYSTRACE_EDEADLK: return (-35);
	case SYSTRACE_ENAMETOOLONG: return (-36);
	case SYSTRACE_ENOLCK: return (-37);
	case SYSTRACE_ENOSYS: return (-38);
	case SYSTRACE_ENOTEMPTY: return (-39);
	case SYSTRACE_ELOOP: return (-40);
	case SYSTRACE_EREMOTE: return (-66);
	case SYSTRACE_EUSERS: return (-87);
	case SYSTRACE_ENOTSOCK: return (-88);
	case SYSTRACE_EDESTADDRREQ: return (-89);
	case SYSTRACE_EMSGSIZE: return (-90);
	case SYSTRACE_EPROTOTYPE: return (-91);
	case SYSTRACE_ENOPROTOOPT: return (-92);
	case SYSTRACE_EPROTONOSUPPORT: return (-93);
	case SYSTRACE_ESOCKTNOSUPPORT: return (-94);
	case SYSTRACE_EOPNOTSUPP: return (-95);
	case SYSTRACE_EPFNOSUPPORT: return (-96);
	case SYSTRACE_EAFNOSUPPORT: return (-97);
	case SYSTRACE_EADDRINUSE: return (-98);
	case SYSTRACE_EADDRNOTAVAIL: return (-99);
	case SYSTRACE_ENETDOWN: return (-100);
	case SYSTRACE_ENETUNREACH: return (-101);
	case SYSTRACE_ENETRESET: return (-102);
	case SYSTRACE_ECONNABORTED: return (-103);
	case SYSTRACE_ECONNRESET: return (-104);
	case SYSTRACE_ENOBUFS: return (-105);
	case SYSTRACE_EISCONN: return (-106);
	case SYSTRACE_ENOTCONN: return (-107);
	case SYSTRACE_ESHUTDOWN: return (-108);
	case SYSTRACE_ETOOMANYREFS: return (-109);
	case SYSTRACE_ETIMEDOUT: return (-110);
	case SYSTRACE_ECONNREFUSED: return (-111);
	case SYSTRACE_EHOSTDOWN: return (-112);
	case SYSTRACE_EHOSTUNREACH: return (-113);
	case SYSTRACE_EALREADY: return (-114);
	case SYSTRACE_EINPROGRESS: return (-115);
	case SYSTRACE_ESTALE: return (-116);
	case SYSTRACE_EDQUOT: return (-122);
	default: return (-1);
	};
}

static void
linux_abortsyscall(pid_t pid)
{
	struct user_regs_struct regs;
	int res = ptrace(PTRACE_GETREGS, pid, NULL, &regs);
	if (res == -1) {
		/* We might have killed the process in the mean time */
		if (errno == ESRCH)
			return;
		err(1, "%s: ptrace getregs", __func__);
	}
	SYSCALL_NUM(&regs) = 0xbadca11;
	res = ptrace(PTRACE_SETREGS, pid, NULL, &regs);
	if (res == -1)
		err(1, "%s: ptrace getregs", __func__);
}

static void
linux_write_returncode(pid_t pid, struct user_regs_struct* regs, long code)
{
	int res;
	DFPRINTF((stderr, "%s: setting return code to %d\n", __func__, code));
	SET_RETURN_CODE(regs, code);
	res = ptrace(PTRACE_SETREGS, pid, NULL, regs);
	if (res == -1)
		err(1, "%s: ptrace getregs", __func__);
}

/*
 * Returns -1 if the syscal needs to be terminated.  Result contains the Linux
 * mapped return value.
 */

static int
linux_policytranslate(int policy, int errnumber, long *result)
{
	long error_code = 0;
	int res = 0;

	DFPRINTF((stderr, "%s: policy %d\n", __func__, policy));
	
	switch (policy) {
	case ICPOLICY_ASK:
	case ICPOLICY_PERMIT:
		/* Not clear that we need to do anything here */
		break;
	case ICPOLICY_NEVER:
		res = -1;
		break;
	default:
		errx(1, "%s: bad policy: %d", __func__, policy);
		break;
	}

	if (res == -1)
		error_code = linux_translate_errno(errnumber);
	
	*result = error_code;
	return (res);
}

/*
 * Aborts a system call with the specified error code.  0 is a valid
 * error code to return.
 */

static void
linux_set_returncode(pid_t pid, long error_code)
{
	struct intercept_pid *icpid;
	struct linux_data *data;
	DFPRINTF((stderr, "%s: setting return code: pid %d error %ld\n",
		__func__, pid, error_code));

	icpid = intercept_findpid(pid);
	if (icpid == NULL)
		err(1, "%s: intercept_getpid", __func__);
	data = icpid->data;
	data->error_code = error_code;
	data->flags |= SYSTR_FLAGS_ERRORCODE;
}

static void
linux_abortsyscall_error(pid_t pid, long error_code)
{
	DFPRINTF((stderr, "%s: aborting system call: pid %d error %d\n",
		__func__, pid, error_code));

	linux_set_returncode(pid, error_code);

	/* Abort the system call */
	linux_abortsyscall(pid);
}

static int
linux_answer(int fd, pid_t pid, u_int32_t seqnr, short policy, int errnumber,
    short flags, struct elevate *elevate)
{
	int res = -1;
	long error_code = 0;
	struct intercept_pid *icpid;
	struct linux_data *data = NULL;

	icpid = intercept_findpid(pid);

	if (icpid == NULL)
		errx(1, "%s: no state for pid %d", __func__, pid);

	data = icpid->data;

	DFPRINTF((stderr, "%s: pid %d action %d errnumber %d\n",
		__func__, pid, policy, errnumber));

	if (flags & ICFLAGS_RESULT)
		data->flags |= linux_translate_flags(flags);

	if (linux_policytranslate(policy, errnumber, &error_code) == -1) {
		linux_abortsyscall_error(pid, error_code);
	} else {
		enum LINUX_CALL_TYPES call_type;
#ifdef PTRACE_LINUX64
		call_type = linux_call_type(data->regs.cs);
#else
		call_type = LINUX32;
#endif
		DFPRINTF((stderr, "%s: allowing system call\n", __func__));
		/* See notes in linux_systemcall(), only set this flag is the fork is
		 * permitted. */
		if (linux_isfork(linux_syscall_name(
					 call_type, pid, SYSCALL_NUM(&data->regs))))
			data->flags |= SYSTR_FLAGS_SAWFORK;
	}
	
	/* we need to deny here if possible */
	res = ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
	if (res == -1) {
		/* It's possible that we just killed the child.  Haha */
		if (errno == ESRCH)
			return (0);
		warn("%s:%d %s: ptrace", __FILE__, __LINE__, __func__);
	}
	
	/* no elevation */

	return (res);
}

static int
linux_newpolicy(int fd)
{
	int i;
	DFPRINTF((stderr, "%s:%d %s: fd %d\n", __FILE__, __LINE__, __func__, fd));

	for (i = 0; i < MAX_POLICIES; i++) {
		if (!policy_used[i])
			break;
	}

	if (i >= MAX_POLICIES) {
		warnx("%s:%d %s: out of policies", __FILE__, __LINE__, __func__);
		return (-1);
	}

	policy_used[i] = 1;

	return (i);
}

static int
linux_assignpolicy(int fd, pid_t pid, int num)
{
	struct intercept_pid *icpid;
	struct linux_data *data;
	DFPRINTF((stderr, "%s: pid %d\n", __func__, pid));

	icpid = intercept_findpid(pid);
	if (icpid == NULL) {
		warnx("%s:%d %s: cannot find pid %d",
		    __FILE__, __LINE__, __func__, pid);
		return (-1);
	}

	data = icpid->data;
	data->policy = num;
	
	return (0);
}

static int
linux_modifypolicy(int fd, int num, int code, short policy)
{
	DFPRINTF((stderr, "%s: fd %d policy %d: %d\n", __func__, fd, num, code));

	if (num < 0 || num >= MAX_POLICIES || code < 0 || code >= MAX_SYSCALLS)
		errx(1, "%s: bad parameters", __func__);

	if (!policy_used[num]) {
		warnx("%s:%d %s: unused policy number %d",
		    __FILE__, __LINE__, __func__, num);
		return (-1);
	}

	policies[num].error_code[code] = policy;

	return (0);
}

static int
linux_replace(int fd, pid_t pid, u_int16_t seqnr,
    struct intercept_replace *repl)
{
	DFPRINTF((stderr, "%s: pid %d\n", __func__, pid));
	/* Replace is not easily supported by PTRACE */
	
	return (0);
}

static int
linux_io(int fd, pid_t pid, int op, void *addr, u_char *buf, size_t size)
{
	int i = 0;
	union {
		long val;
		char x[sizeof(long)];
	} u;

	DFPRINTF((stderr, "%s: pid %d, %p for %ld\n", __func__, pid, addr, size));

	if (op != INTERCEPT_READ)
		errx(1, "%s: unsupported IO operation", __func__);

	if ((long)addr & (sizeof(long) - 1)) {
		/* the address is not word aligned */
		int off = (long)addr - ((long)addr & -sizeof(long));
		int tocopy;
		addr = (void *)((long)addr & -sizeof(long));
		u.val = ptrace(PTRACE_PEEKDATA, pid, addr, 0);
		if (u.val == -1) {
			DFPRINTF((stderr,
				"%s:%d %s: read at %p failed (start): %d\n",
				__FILE__, __LINE__, __func__, addr, errno));
			errno = EINVAL;
			return (-1);
		}

		tocopy = MIN(sizeof(long) - off, size);
		memcpy(buf, &u.x[off], tocopy);
		addr += sizeof(long);
		buf += tocopy;
		size -= tocopy;
	}
	
	while (i < size) {
		u.val = ptrace(PTRACE_PEEKDATA, pid, addr, 0);
		/* Man page says this is how failure is indicated */
		if (u.val == -1 && errno != 0) {
			DFPRINTF((stderr,
				"%s:%d %s: read at %p failed (%d of %ld): %d\n",
				__FILE__, __LINE__, __func__, addr, i, size, errno));
			errno = EINVAL;
			return (-1);
		}
		memcpy(buf, u.x, size - i < sizeof(long) ? size - i : sizeof(long));

		i += sizeof(long);
		addr = (char *)addr + sizeof(long);
		buf += sizeof(long);
	}
	
	return (0);
}

static int
linux_setcwd(int fd, pid_t pid)
{
	// return (ioctl(fd, STRIOCGETCWD, &pid));
	return (0);
}

static int
linux_restcwd(int fd)
{
	int res = 0;

	// if ((res = ioctl(fd, STRIOCRESCWD, 0)) == -1)
	//	warn("%s: ioctl", __func__); /* XXX */

	return (res);
}

static int
linux_argument(int off, void *pargs, int argsize, void **pres)
{
	struct user_regs_struct *regs = pargs;
	DFPRINTF((stderr, "%s: off %d\n", __func__, off));
	
	switch (off) {
	case 0:
		*pres = (void *)ARGUMENT_0(regs);
		break;
	case 1:
		*pres = (void *)ARGUMENT_1(regs);
		break;
	case 2:
		*pres = (void *)ARGUMENT_2(regs);
		break;
	case 3:
		*pres = (void *)ARGUMENT_3(regs);
		break;
	case 4:
		*pres = (void *)ARGUMENT_4(regs);
		break;
	case 5:
		*pres = (void *)ARGUMENT_5(regs);
		break;
	default:
		/* out of bounds */
		return (-1);
	}

	DFPRINTF((stderr, "%s: off %d: 0x%lx\n", __func__, off, (long)*pres));

	return (0);
}

static int
linux_set_argument(struct user_regs_struct *regs, int off, long val)
{
	DFPRINTF((stderr, "%s: set off %d to %ld\n", __func__, off, val));
	
	switch (off) {
	case 0:
		SET_ARGUMENT_0(regs, val);
		break;
	case 1:
		SET_ARGUMENT_1(regs, val);
		break;
	case 2:
		SET_ARGUMENT_2(regs, val);
		break;
	case 3:
		SET_ARGUMENT_3(regs, val);
		break;
	case 4:
		SET_ARGUMENT_4(regs, val);
		break;
	case 5:
		SET_ARGUMENT_5(regs, val);
		break;
	default:
		/* out of bounds */
		errx(1, "%s: seting bad argument %d", __func__, off);
	}
	

	return (0);
}

static int
linux_lookuppolicy(struct intercept_pid *icpid, int sysnum)
{
	struct linux_data *data = icpid->data;
	
	if (sysnum < 0 || sysnum >= MAX_SYSCALLS)
		errx(1, "%s: bad system call %d", __func__, sysnum);

	if (data->policy < 0 || data->policy >= MAX_POLICIES)
		errx(1, "%s: bad policy %d", __func__, data->policy);

	DFPRINTF((stderr, "%s: lookup policy %d: syscall %d -> %d\n",
		__func__, data->policy,
		sysnum, policies[data->policy].error_code[sysnum]));
	return (policies[data->policy].error_code[sysnum]);
}

/*
 * Forkers need super special treatment
 */

static int
linux_isfork(const char *sysname)
{
	if (sysname == NULL)
		return (0);

	return (!strcmp(sysname, "fork") ||
	    !strcmp(sysname, "vfork") ||
	    !strcmp(sysname, "clone"));
}

/*
 * waitpiders need special super treatment
 */

static int
linux_iswaitpid(const char *sysname)
{
	if (sysname == NULL)
		return (0);
	
	return (!strcmp(sysname, "wait4") || !strcmp(sysname, "waitpid"));
}

/*
 * This is a trick that I learned from Subterfugue.  Lots of Kudos to
 * Mike Coleman.
 */

static void
linux_rewritefork(pid_t pid, const char *sysname, struct user_regs_struct *regs)
{
	int res;

	DFPRINTF((stderr, "%s: pid %d rewriting %s to clone\n",
		__func__, pid, sysname));
	
	SYSCALL_NUM(regs) = linux_map_call(linux_call_type(regs->cs), SYSTR_CLONE);
	if (strcmp(sysname, "fork") == 0) {
		int clone_flags = SIGCHLD | CLONE_PTRACE;
		linux_set_argument(regs, 0, clone_flags);
		linux_set_argument(regs, 1, 0);
	} else if (strcmp(sysname, "vfork") == 0) {
		int clone_flags = SIGCHLD | CLONE_PTRACE | CLONE_VFORK;
		linux_set_argument(regs, 0, clone_flags);
		linux_set_argument(regs, 1, 0);
	} else {
		/* clone */
		long clone_flags;
		linux_argument(0, regs, sizeof(*regs), (void **)&clone_flags);
		if ((clone_flags & CLONE_PTRACE) == 0) {
			clone_flags |= CLONE_PTRACE;
			linux_set_argument(regs, 0, clone_flags);
		}
	}
	res = ptrace(PTRACE_SETREGS, pid, NULL, regs);
	if (res == -1)
		err(1, "%s: ptrace getregs", __func__);
}

static int
linux_havewaitchildren(
	struct intercept_pid *icpid,
	struct linux_data *data,
	pid_t wait_pid)
{
	if (data->flags & SYSTR_FLAGS_CLONE_THREAD) {
		icpid = intercept_findpid(icpid->ppid);
		if (icpid == NULL)
			errx(1, "%s: interept_find", __func__);
		data = icpid->data;
	}

	/* if there are no children around - we can return immediately */
	if (data->nchildren <= data->nthreads_detached)
		return (0);
	
	if (wait_pid > 0) {
		struct intercept_pid *npid = intercept_findpid(wait_pid);
		if (npid == NULL || npid->ppid != icpid->pid) {
			/*
			 * the child is not around any longer or this
			 * is not the right parent.  The system call
			 * should return ECHLD - so we just let it
			 * continue.
			 */

			return (0);
		}
	}

	return (1);
}

static void
linux_rewritewaitpid(pid_t pid, const char *sysname,
    struct user_regs_struct *regs)
{
	struct intercept_pid *icpid;
	struct linux_data *data;
	int res;
	long wait_pid;
	long woptions;
	long orig_eax = SYSCALL_NUM(regs);

	DFPRINTF((stderr, "%s: pid %d rewriting %s\n", __func__, pid, sysname));
	
	linux_argument(2, regs, sizeof(*regs), (void **)&woptions);

	/* check the arguments for validity */
	if (woptions & ~(WNOHANG|WUNTRACED|__WCLONE|__WALL)) {
		/* probably not needed but good for completeness */
		linux_abortsyscall_error(pid, -EINVAL);
		return;
	}

	icpid = intercept_findpid(pid);
	if (icpid == NULL)
		errx(1, "%s: intercept_findpid %d", __func__, pid);
	data = icpid->data;

	/* See if the process is waiting for something */
	linux_argument(0, regs, sizeof(*regs), (void **)&wait_pid);
	linux_argument(1, regs, sizeof(*regs), (void **)&data->pstatus);
	DFPRINTF((stderr, "%s: pid %d waitpid on %ld\n",
		__func__, pid, wait_pid));
	data->waitpid = linux_find_pidstatus(icpid, wait_pid, NULL);
	if (data->waitpid != -1) {
		/* Mark this system call as pending in waitpid */
		data->flags |= SYSTR_FLAGS_SAWWAITPID;

		/* We have data to report - weeh */
		/*
		 * Allow the real system call to continue to reap children,
		 * that got reparented to the process when it died. We are
		 * still going to rewrite the return value and hope that
		 * everything goes well.
		 */
		return;
	}

	if (woptions & WNOHANG) {
		/* if we have nothing to report, we just return immediately */
		DFPRINTF((stderr, "%s: pid %d wait4 returning nothing\n",
			__func__, pid));

		/*
		 * allow the system call but make sure that it does not
		 * return anything useful.
		 */
		linux_set_returncode(pid, 0);
		return;
	}

	data->waitpid = wait_pid;
	if (!linux_havewaitchildren(icpid, data, wait_pid)) {
		/* make waitpid return immediately */
		woptions |= WNOHANG;
		linux_set_argument(regs, 2, woptions);
		res = ptrace(PTRACE_SETREGS, pid, NULL, regs);
		if (res == -1)
			err(1, "%s: ptrace getregs", __func__);

		/*
		 * There are no children to wait for -> ECHILD
		 */
		linux_set_returncode(pid, -ECHILD);
		return;
	}
		
	/* treat properly */
	data->flags |= SYSTR_FLAGS_PAUSING;

	/* if this thread is a cloner, then we need to mark waiting children */
	if (data->flags & SYSTR_FLAGS_CLONE_THREAD) {
		struct intercept_pid *ppid = intercept_findpid(icpid->ppid);
		struct linux_data *pdata;
		if (ppid == NULL)
			errx(1, "%s: intercept_findpid", __func__);
		pdata = ppid->data;
		pdata->nthreads_waiting++;
	}
		
	/* XXX - turn it into a pause */
	DFPRINTF((stderr, "%s: pid %d wait4 pausing\n",	__func__, pid));
	SYSCALL_NUM(regs) = linux_map_call(linux_call_type(regs->cs), SYSTR_GETPID);
       
	res = ptrace(PTRACE_SETREGS, pid, NULL, regs);
	if (res == -1)
		err(1, "%s: ptrace getregs", __func__);

	/*
	 * So that the policy matching still works - let's hope that
	 * nobody is going to write to the registers after us.
	 */
	SYSCALL_NUM(regs) = orig_eax;

}

static void
linux_rewritewaitpid_return(pid_t pid, struct linux_data *data, pid_t res_pid, int res_status)
{
	struct user_regs_struct regs;
	int res;

	DFPRINTF((stderr, "%s: pid %d returning from wait4: %d %lx\n",
		__func__, pid, data->waitpid, data->pstatus));

	res = ptrace(PTRACE_GETREGS, pid, NULL, &regs);
	if (res == -1)
		err(1, "%s: ptrace getregs: pid %d", __func__, pid);

	linux_write_returncode(pid, &regs, res_pid);
	data->waitpid = -1;

	/* Stick status */
	if (data->pstatus != 0) {
		res = ptrace(PTRACE_POKEDATA, pid,
		    (void *)data->pstatus, (void *)(0L | res_status));
		if (res == -1)
			err(1, "%s: pokeuser pid %d: %x",
			    __func__, pid, data->pstatus);
	}

	/*
	 * We also need to make sure that the system call is the correct
	 * one again.  In case that there is a restart, we do not want to
	 * enter a pause.
	 */
	SYSCALL_NUM(&regs) = linux_map_call(linux_call_type(regs.cs), SYSTR_WAIT4);

	res = ptrace(PTRACE_SETREGS, pid, NULL, &regs);
	if (res == -1)
		err(1, "%s: ptrace getregs", __func__);

	/* We make this an undeniable system call */
	data->flags &= ~SYSTR_FLAGS_ERRORCODE;
}

static void
linux_skipsigstop(pid_t pid)
{
	struct intercept_pid *icpid;
	struct linux_data *data;
	icpid = intercept_findpid(pid);
	if (icpid == NULL)
		err(1, "%s: intercept_findpid", __func__);
	data = icpid->data;
	data->flags |= SYSTR_FLAGS_SKIPSTOP;
}

/* we got a new thread/process - keep track of meta data */
static void
linux_child_info(pid_t pid, pid_t cpid, struct user_regs_struct *regs)
{
	struct intercept_pid *icpid, *cicpid = NULL;
	struct linux_data *data, *cdata = NULL;
	long clone_flags;

	/* get the meta-data for the parent pid  */
	icpid = intercept_findpid(pid);
	if (icpid == NULL)
		err(1, "%s: intercept_findpid", __func__);
	data = icpid->data;

	/* figure out what kind of clone this was */
	linux_argument(0, regs, sizeof(*regs), (void **)&clone_flags);

	/* get the meta-data for the child pid if we need it later  */
	if ((clone_flags & CLONE_THREAD) ||
	    (data->flags & SYSTR_FLAGS_CLONE_THREAD)) {
		cicpid = intercept_findpid(cpid);
		if (cicpid == NULL)
			err(1, "%s: intercept_findpid", __func__);
		cdata = cicpid->data;
	}
	
	DFPRINTF((stderr, "%s: pid %d (parent %d) clone to %d(%d): %lx\n",
		__func__, pid, icpid->ppid, cpid,
		cdata != NULL ? cdata->nchildren : 0, clone_flags));
	
	if (data->flags & SYSTR_FLAGS_CLONE_THREAD) {
		/*
		 * if we are a thread already then we need to attribute
		 * the child to the parent of this thread.
		 */
		pid_t ppid = icpid->ppid;
		if ((icpid = intercept_findpid(ppid)) == NULL)
			errx(1, "%s: pid %d cannot find parent %d",
			    __func__, pid, ppid);

		/* count the children with the parent */
		data = icpid->data;

		/* make this child belong to the parent of the current pid */
		cicpid->ppid = ppid;

		DFPRINTF((stderr, "%s: set parent for %d to %d\n",
			__func__, cpid, ppid));
	}

	/*
	 * count the children with the correct parent - which is either
	 * pid if we were not a thread or our parent.
	 */
	data->nchildren++;
	DFPRINTF((stderr, "%s: pid %d: Increasing child count to %d\n",
		__func__, icpid->pid, data->nchildren));

	if (clone_flags & CLONE_THREAD) {
		cdata->flags |= SYSTR_FLAGS_CLONE_THREAD;
		data->nthreads++;
	}

	if (clone_flags & CLONE_DETACHED) {
		cdata->flags |= SYSTR_FLAGS_CLONE_DETACHED;
		data->nthreads_detached++;
	}
}

static void
linux_forkreturn(pid_t pid, struct user_regs_struct *regs)
{
	int sysnum = SYSCALL_NUM(regs);
	pid_t child_pid = RETURN_CODE(regs);
	const char *sysname;
	enum LINUX_CALL_TYPES call_type;
#ifdef PTRACE_LINUX64
	call_type = linux_call_type(regs->cs);
#else
	call_type = LINUX32;
#endif
	sysname = linux_syscall_name(call_type, pid, sysnum);

	if (!linux_isfork(sysname)) {
		/* should only be called on fork() return */
		errx(1, "%s: %s was not expected, i should not have been called.",
			__func__, sysname);
	}

	DFPRINTF((stderr, "%s: pid %d fork return %d\n",
		__func__, pid, child_pid));
	if (child_pid >= 0) {
		struct linux_data *cdata = NULL;
		int child_insigstop = 0;
		/* a clone returned successfully */
		if (!child_pid) {
			/* something funky?? */
			errx(1, "%s: funky on clone return", __func__);
		}

		/* register a new child with our tracer */
		DFPRINTF((stderr, "%s: pid %d -> new child %d\n",
			__func__, pid, child_pid));
		/* check if the child is already there and if we kept it in sigtop */
		cdata = linux_get_piddata(child_pid);
		if (cdata != NULL && (cdata->flags & SYSTR_FLAGS_STOPWAITING))
		  child_insigstop = 1;
		intercept_child_info(pid, child_pid);
		linux_child_info(pid, child_pid, regs);
		if (!child_insigstop) {
		  /* we are still expecting a sigstop */
		  linux_skipsigstop(child_pid);
		} else {
		  /* the child is currently waiting in sigstop - continue it */
		  int res = ptrace(PTRACE_SYSCALL, child_pid, (char *)1, 0);
		  if (res == -1)
		    err(1, "%s: ptrace", __func__);
		}
	}
}

static void
linux_setsidreturn(pid_t pid, struct user_regs_struct *regs)
{
	struct intercept_pid *icpid = intercept_findpid(pid);
	struct linux_data *data;
	
	DFPRINTF((stderr, "%s: pid %d setsid return\n", __func__, pid));
	if (icpid == NULL)
		err(1, "%s: intercept_findpid", __func__);
	data = icpid->data;

	data->pgid = pid;
}

static void
linux_setpgidreturn(pid_t pid, struct user_regs_struct *regs)
{
	struct intercept_pid *icpid;
	struct linux_data *data;
	long set_pid, tmp;
	
	linux_argument(0, regs, sizeof(*regs), (void **)&set_pid);
	DFPRINTF((stderr, "%s: pid %d setpgid return: target pid %d\n",
		__func__, pid, set_pid));
	
	icpid = intercept_findpid(pid);
	if (icpid == NULL)
		err(1, "%s: intercept_findpid", __func__);
	data = icpid->data;

	/* remember the new pgid */
	linux_argument(1, regs, sizeof(*regs), (void **)&tmp);
	data->pgid = tmp;
	if (data->pgid == 0)
		data->pgid = pid;
}

static void
linux_systemcall(int fd, pid_t pid, struct intercept_pid *icpid)
{
	// System call intercepted
	struct linux_data *data = NULL;
	struct user_regs_struct *regs;
	const char *sysname = NULL;
	int sysnumber = -1;
	int res;
	enum LINUX_CALL_TYPES call_type;

	data = icpid->data;
	regs = &data->regs;

	if (data->status == SYSCALL_START) {
		res = ptrace(PTRACE_GETREGS, pid, NULL, regs);
		if (res == -1)
			err(1, "%s: ptrace getregs", __func__);
	} else {
		struct user_regs_struct tmp;
		/* Just get the return code */
		res = ptrace(PTRACE_GETREGS, pid, NULL, &tmp);
		if (res == -1)
			err(1, "%s: ptrace getregs", __func__);
		SET_RETURN_CODE(regs, RETURN_CODE(&tmp));
	}

#ifdef PTRACE_LINUX64
	call_type = linux_call_type(regs->cs);
#else
	call_type = LINUX32;
#endif

	sysnumber = SYSCALL_NUM(regs);
	sysname = linux_syscall_name(call_type, pid, sysnumber);

	DFPRINTF((stderr, "%s: pid %d %s(%ld) %s %ld\n",__func__, pid,
		sysname, sysnumber,
		data->status == SYSCALL_START ? "start" : "end",
		(long)-RETURN_CODE(regs)));

	if (data->status == SYSCALL_START) {
#ifndef PTRACE_LINUX64
		if (regs->cs != 0x73 && regs->cs != 0x23) {
			/*
			 * Security violation - attacker might be trying to
			 * map to the 64-bit syscall table.
			 */
			linux_abortsyscall(pid);
			errx(1, "%s: evil CS value 0x%x", __func__, regs->cs);
		}
#endif
		if (sysnumber == -1) {
			/* Spurious stuff - ignore? */
			DFPRINTF((stderr, "%s: spurious -1 on syscall name\n",
				__func__));
			ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
			return;
		}

		if (-RETURN_CODE(regs) != ENOSYS) {
		  /* 
		   * On some Linux system's it's possible for the
		   * child to return before we get the SIGSTOP.  That
		   * also means that the child may return before the
		   * parent does. It's pretty weird.
		   */
		  if ((data->flags & SYSTR_FLAGS_SKIPSTOP) ||
		      data->policy == -1) {
		    DFPRINTF((stderr, "%s: pid %d forcing sys call exit\n",
			      __func__, pid));
		    data->status = SYSCALL_END;
		  } else {
		    errx(1, "%s: got system call start without ENOSYS", __func__);
		  }
		}
	}
	
	if (data->status == SYSCALL_START) {
		char *emulation = "linux";
		int policy;
		long error_code;
		
		data->status = SYSCALL_END;

		if (linux_isfork(sysname)) {
			/*
			 * Make everything a clone - and tell clone to ptrace
			 * the children for us.
			 */
			linux_rewritefork(pid, sysname, regs);
		} else if (linux_iswaitpid(sysname)) {
			linux_rewritewaitpid(pid, sysname, regs);
		}
		
		if (data->policy != -1) {
			/* Check if we should use the fast path */
			policy = linux_lookuppolicy(icpid, sysnumber);
			if (policy == ICPOLICY_PERMIT) {
				if (linux_isfork(sysname)) {
					/* 
					 * At this point we know we've
					 * seen a fork()-like call,
					 * and it is going to be
					 * permitted. Setting this
					 * flag indicates that on the
					 * next syscall return, we
					 * need to call
					 * linux_forkreturn() to
					 * handle setting up a new
					 * child process.
					 *
					 * It's important to only set
					 * this flag if we really
					 * expect the next syscall
					 * return to be from a fork(),
					 * although it could have
					 * failed due to erroneous
					 * clone flags or process
					 * limits, etc. which
					 * linux_forkreturn() must
					 * handle.
					 *
					 * This flag can also be set in linux_answer().
					 */
					data->flags |= SYSTR_FLAGS_SAWFORK;
				}
				res = ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
				if (res == -1)
					err(1, "%s: ptrace getregs", __func__);
				return;
			} else if (linux_policytranslate(
					   policy >= 1, policy, &error_code) == -1) {
				linux_abortsyscall_error(pid, error_code);
				res = ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
				if (res == -1)
					err(1, "%s: ptrace getregs", __func__);
				return;
			}
		}

#ifdef PTRACE_LINUX64
		if (call_type == LINUX64)
			emulation = "linux64";
#endif
		intercept_syscall(fd, pid, 1, data->policy,
		    sysname, sysnumber, emulation,
		    (void *)regs, sizeof(*regs));
	} else {
		char *emulation = "linux";

		/* System call return */
		data->status = SYSCALL_START;

		/* Check if the system call should be paused */
		if (data->flags & SYSTR_FLAGS_PAUSING) {
			DFPRINTF((stderr, "%s: leaving pid %d paused\n",
				__func__, pid));
			return;
		}
		
		/* first stick the regular results */
		if (data->flags & SYSTR_FLAGS_SAWWAITPID) {
			pid_t res_pid;
			int res_status;
			data->flags &= ~SYSTR_FLAGS_SAWWAITPID;
			/* set the result for the immediate return */
			res_pid = linux_find_pidstatus(icpid, data->waitpid,
			    &res_status);
			linux_rewritewaitpid_return(pid, data,
			    res_pid, res_status);
			linux_remove_pidstatus(icpid, res_pid);
		}

		/* it's still possible that we got another abort */
		if (data->flags & SYSTR_FLAGS_ERRORCODE) {
			linux_write_returncode(pid, regs, data->error_code);
			data->flags &= ~SYSTR_FLAGS_ERRORCODE;
			data->error_code = 0;
		} else if (sysnumber == linux_map_call(call_type, SYSTR_EXECVE) &&
		    RETURN_CODE(regs) == 0) {
			/* remember that we saw a successful execve */
			data->flags |= SYSTR_FLAGS_SAWEXECVE;
		} else if (data->flags & SYSTR_FLAGS_SAWFORK) {
			data->flags &= ~SYSTR_FLAGS_SAWFORK;
			linux_forkreturn(pid, regs);
		} else if (sysnumber == linux_map_call(call_type, SYSTR_SETSID) &&
		    RETURN_CODE(regs) >= 0) {
			linux_setsidreturn(pid, regs);
		} else if (sysnumber == linux_map_call(call_type, SYSTR_SETPGID) &&
		    RETURN_CODE(regs) == 0) {
			linux_setpgidreturn(pid, regs);
		}
		
		/* We did not want result interception */
		if ((data->flags & SYSTR_FLAGS_RESULT) == 0) {
			DFPRINTF((stderr, "%s: continue pid %d\n",
				__func__, pid));
			res = ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
			if (res == -1)
				err(1, "%s: ptrace getregs", __func__);
			return;
		}

		data->flags &= ~SYSTR_FLAGS_RESULT;

#ifdef PTRACE_LINUX64
		if (linux_call_type(regs->cs) == LINUX64)
			emulation = "linux64";
#endif
		intercept_syscall_result(fd, pid, 1, data->policy,
		    sysname, sysnumber, emulation,
		    (void *)regs, sizeof(*regs),
		    -RETURN_CODE(regs), 0 /* rval */);
	}
}

static void
linux_resumeparent(struct intercept_pid *icpid)
{
	struct linux_data *data = icpid->data;
	DFPRINTF((stderr, "%s: resuming pid %d\n", __func__, icpid->pid));

	if (!(data->flags & SYSTR_FLAGS_PAUSING))
		errx(1, "%s: pid %d is not pausing\n", __func__, icpid->pid);

	data->flags &= ~SYSTR_FLAGS_PAUSING;
	/* let the parent know that we are no longer waiting */
	if (data->flags & SYSTR_FLAGS_CLONE_THREAD) {
		struct intercept_pid *ppid = intercept_findpid(icpid->ppid);
		struct linux_data *pdata;
		if (ppid == NULL)
			errx(1, "%s: intercept_findpid", __func__);
		pdata = ppid->data;
		pdata->nthreads_waiting--;
	}

	ptrace(PTRACE_SYSCALL, icpid->pid, (char *)1, 0);
}

/*
 * Wakes all processes that are potentially waiting on this child to return.
 */

struct linux_search_state {
	pid_t ppid;
	struct intercept_pid** pids;
	int offset;
	int size;
};

static void
linux_wakeprocess_fill(struct intercept_pid *icpid, void *arg)
{
	struct linux_search_state *state = arg;
	struct linux_data *data = icpid->data;
	
	if (icpid->ppid != state->ppid)
		return;

	/* only return threads */
	if ((data->flags & SYSTR_FLAGS_CLONE_THREAD) == 0)
		return;

	/* and only if they are paused */
	if ((data->flags & SYSTR_FLAGS_PAUSING) == 0)
		return;
	
	if (state->offset >= state->size)
		errx(1, "%s: overflow", __func__);
	state->pids[state->offset++] = icpid;
}

static void
linux_wakeprocesses(struct intercept_pid *icpid, pid_t wpid)
{
	struct linux_data *data = icpid->data;
	pid_t res_pid;
	int res_status;
	int resumed = 0;
	
	DFPRINTF((stderr, "%s: trying to wake up pid %d\n",
		__func__, icpid->pid));

	if (data->flags & SYSTR_FLAGS_PAUSING) {
		res_pid = linux_find_pidstatus(icpid, data->waitpid,
		    &res_status);
		if (res_pid != -1) {
			/*
			 * Wake the waiting process up and hope that
			 * it gets to pick up the correct wait status.
			 */
			linux_rewritewaitpid_return(icpid->pid, data,
			    res_pid, res_status);
			linux_resumeparent(icpid);
			resumed++;
		}
	}

	if (data->nthreads_waiting > 0) {
		struct linux_search_state state;
		int i;

		state.ppid = icpid->pid;
		state.offset = 0;
		state.size = data->nthreads_waiting;
		state.pids = malloc(state.size * sizeof(struct intercept_pid *));
		intercept_foreachpid(linux_wakeprocess_fill, &state);

		/* do the direct waiters */
		for (i = 0; i < state.offset; ++i) {
			struct intercept_pid *cpid = state.pids[i];
			struct linux_data *cdata = cpid->data;

			if (cdata->waitpid != wpid)
				continue;

			res_pid = linux_find_pidstatus(cpid, wpid, &res_status);
			if (res_pid == -1)
				errx(1, "%s: got bad pid return", __func__);
			linux_rewritewaitpid_return(cpid->pid, cdata,
			    res_pid, res_status);
			linux_resumeparent(cpid);
			resumed++;
		}

		/* do the non-direct waiters */
		if (resumed == 0) {
			for (i = 0; i < state.offset; ++i) {
				struct intercept_pid *cpid = state.pids[i];
				struct linux_data *cdata = cpid->data;

				if (cdata->waitpid == wpid)
					continue;

				if ((res_pid = linux_find_pidstatus(cpid,
					    cdata->waitpid,
					    &res_status)) == -1)
					continue;
				
				linux_rewritewaitpid_return(cpid->pid, cdata,
				    res_pid, res_status);
				linux_resumeparent(cpid);
				resumed++;
			}
		}
	}

	/* Remove the status for the child that caused the death */
	if (resumed)
		linux_remove_pidstatus(icpid, wpid);
}

static void
linux_kill(struct intercept_pid *icpid, void* arg)
{
        ptrace(PTRACE_KILL, icpid->pid, NULL, NULL);
}

static void
linux_atexit(void)
{
        /* Kill all of the traced PIDs. */
        intercept_foreachpid(&linux_kill, NULL);
}

static int
linux_read(int fd)
{
	// here is where all the magic happens
	struct intercept_pid *icpid;
	struct linux_data *data;
	int status, res;
	pid_t pid;

	do {
		DFPRINTF((stderr, "%s: waiting on syscall\n", __func__));
		pid = waitpid(-1, &status, WUNTRACED | __WALL);
		
	} while (pid == -1 && errno == EINTR);

	DFPRINTF((stderr, "%s: status %d pid %d\n", __func__, status, pid));

	if (pid == -1)
		err(1, "%s: wait", __func__);

	icpid = linux_getpid(pid);
	if (icpid == NULL) {
		warnx("%s:%d %s: cannot find pid %d",
		    __FILE__, __LINE__, __func__, pid);
		return (-1);
	}

	if (WIFEXITED(status)) {
		linux_childdead(pid, status);
	} else if (WIFSTOPPED(status) && WSTOPSIG(status) == sigsyscall) {
		/* This test means that PTRACE_O_TRACESYSGOOD failed */
		if (sigsyscall == SIGTRAP) {
			/*
			 * Apparently, we get a SIGTRAP after an
			 * execve.  However, without
			 * PTRACE_O_TRACESYSGOOD, that fact is hidden
			 * to us.
			 */
			data = icpid->data;
			if (data->flags & SYSTR_FLAGS_SAWEXECVE) {
				/* Linux is a weird beast */
				data->flags &= ~SYSTR_FLAGS_SAWEXECVE;
				ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
				return (0);
			}
		}
		linux_systemcall(fd, pid, icpid);
	} else if (WIFSTOPPED(status) && WSTOPSIG(status) != sigsyscall) {
		int signum = WSTOPSIG(status);
		/*
		 * Apparently, we get a SIGTRAP after an execve, even if
		 * we specified the PTRACE_O_TRACESYSGOOD option.
		 */
		data = icpid->data;
		if (signum == SIGTRAP &&
		    (data->flags & SYSTR_FLAGS_SAWEXECVE)) {
			/* Linux is a weird beast */
			data->flags &= ~SYSTR_FLAGS_SAWEXECVE;
			ptrace(PTRACE_SYSCALL, pid, (char *)1, 0);
			return (0);
		}

		/*
		 * New childs may get a gratutious skip stop for being
		 * attached to the tracing facility.  Ignore it and make
		 * them continue to run.
		 */
		if (signum == SIGSTOP) {
			if (data->flags & SYSTR_FLAGS_SKIPSTOP) {
				data->flags &= ~SYSTR_FLAGS_SKIPSTOP;
				signum = 0;
				DFPRINTF((stderr,
					  "%s: making new child %d continue\n",
					  __func__, pid));
			} else if (data->policy == -1) {
			  /* 
			   * We are not going to wake this child up, until we
			   * header back from our parent.
			   */
			  data->flags |= SYSTR_FLAGS_STOPWAITING;
			  return (0);
			}
		}

		if (signum != 0)
			DFPRINTF((stderr, "%s: passing signal %d to %d\n",
				  __func__, signum, pid));
		res = ptrace(PTRACE_SYSCALL, pid, NULL, signum);
		if (res == -1)
			err(1, "%s: ptrace signal passthrough", __func__);
	} else if (WIFSIGNALED(status)) {
		linux_childdead(pid, status);
	} else {
		errx(1, "%s: unhandled waitpid case", __func__);
	}

	/* bad hack to propagate exit information */
	if (!intercept_existpids()) {
		extern int systrace_dumppolicies(int);
		DFPRINTF((stderr, "%s: Exiting with %d",
			__func__, WEXITSTATUS(status)));
		systrace_dumppolicies(1);
		exit(WEXITSTATUS(status));
	}
	
	return (0);
}

struct intercept_system intercept = {
#ifdef PTRACE_LINUX64
	"linux64 ptrace",
#else
	"linux32 ptrace",
#endif
	linux_init,
	linux_open,
	linux_attach,
	linux_detach,
	linux_report,
	linux_read,
	linux_syscall_number,
	linux_setcwd,
	linux_restcwd,
	linux_io,
	linux_argument,
	linux_answer,
	linux_newpolicy,
	linux_assignpolicy,
	linux_modifypolicy,
	linux_replace,
	linux_clonepid,
	linux_freepid,
};
