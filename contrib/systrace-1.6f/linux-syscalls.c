/*
 * Copyright (c) 2002 Marius Aamodt Eriksen <marius@umich.edu>
 * Copyright (c) 2002 Niels Provos <provos@citi.umich.edu>
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
#include <sys/ioctl.h>

typedef u_int32_t u32;

#include <asm/unistd.h>

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif /* HAVE_CONFIG_H */

#include <linux/limits.h>
#include <linux/types.h>
#include <linux/systrace.h>
#include <sys/queue.h>
#include <sys/tree.h>

#ifndef NR_syscalls
#define NR_syscalls 512
#endif

#include <limits.h>
#include <err.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

#include "intercept.h"
#include "linux_syscalls.c"
#include "systrace-errno.h"

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
static short                 linux_translate_policy(short);
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

static int
linux_init(void)
{
	return (0);
}

static int
linux_attach(int fd, pid_t pid)
{
	return (ioctl(fd, STRIOCATTACH, &pid) == -1 ? -1 : 0);
}

static int
linux_report(int fd, pid_t pid)
{
	return (0);
}

static int
linux_detach(int fd, pid_t pid)
{
	return (ioctl(fd, STRIOCDETACH, &pid) == -1 ? -1 : 0);
}

static int
linux_open(void)
{
	char *path = "/dev/systrace";
	int fd;

	if ((fd = open(path, O_RDWR, 0)) == -1) {
		warn("%s:%d: open: %s", __FILE__, __LINE__, path);
		return (-1);
	}

	if (fcntl(fd, F_SETFD, 1) == -1)
		warn("%s:%d: fcntl(F_SETFD)", __FILE__, __LINE__);

	return (fd);
}

static struct intercept_pid *
linux_getpid(pid_t pid)
{
	struct intercept_pid *icpid;

	icpid = intercept_getpid(pid);
	if (icpid == NULL)
		return (NULL);

	/* no data to attach yet */

	return (icpid);
}

static void
linux_freepid(struct intercept_pid *ipid)
{
}

static void
linux_clonepid(struct intercept_pid *opid, struct intercept_pid *npid)
{
}

static short
linux_translate_policy(short policy)
{
	switch (policy) {
	case ICPOLICY_ASK:
		return (SYSTR_POLICY_ASK);
	case ICPOLICY_PERMIT:
		return (SYSTR_POLICY_PERMIT);
	case ICPOLICY_NEVER:
		return (SYSTR_POLICY_NEVER);
	default:
		if (policy > 0)
			return (-policy);
		return (SYSTR_POLICY_NEVER);
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
linux_translate_errno(int errno)
{
	/* XXX kind of nasty; make a table instead.  at least it was
	 * automatically generated. */

	switch (errno) {
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


	int (*answer)(int, pid_t, u_int32_t, short, int, short,
	    struct elevate *);
static int
linux_answer(int fd, pid_t pid, u_int32_t seqnr, short policy, int errno,
    short flags, struct elevate *elevate)
{
	struct systrace_answer ans;

	ans.stra_pid = pid;
	ans.stra_seqnr = seqnr;
	ans.stra_policy = linux_translate_policy(policy);
	ans.stra_flags = linux_translate_flags(flags);
	ans.stra_error = linux_translate_errno(errno);

	if (elevate != NULL) {
		if (elevate->e_flags & ELEVATE_UID) {
			ans.stra_flags |= SYSTR_FLAGS_SETEUID;
			ans.stra_seteuid = elevate->e_uid;
		}
		if (elevate->e_flags & ELEVATE_GID) {
			ans.stra_flags |= SYSTR_FLAGS_SETEGID;
			ans.stra_setegid = elevate->e_gid;
		}
	}

	return (ioctl(fd, STRIOCANSWER, &ans) == -1 ? -1 : 0);
}

static int
linux_newpolicy(int fd)
{
	struct systrace_policy pol;

	pol.strp_op = SYSTR_POLICY_NEW;
	pol.strp_num = -1;
	pol.strp_maxents = NR_syscalls;

	return (ioctl(fd, STRIOCPOLICY, &pol) == -1 ? -1 : pol.strp_num);
}

static int
linux_assignpolicy(int fd, pid_t pid, int num)
{
	struct systrace_policy pol;

	pol.strp_op = SYSTR_POLICY_ASSIGN;
	pol.strp_num = num;
	pol.strp_pid = pid;

	return (ioctl(fd, STRIOCPOLICY, &pol) == -1 ? -1 : 0);
}

static int
linux_modifypolicy(int fd, int num, int code, short policy)
{
	struct systrace_policy pol;

	pol.strp_op = SYSTR_POLICY_MODIFY;
	pol.strp_num = num;
	pol.strp_code = code;
	pol.strp_policy = linux_translate_policy(policy);

	return (ioctl(fd, STRIOCPOLICY, &pol) == -1 ? -1 : 0);
}

static int
linux_replace(int fd, pid_t pid, u_int16_t seqnr,
    struct intercept_replace *repl)
{
	struct systrace_replace replace;
	size_t len, off;
	int i, ret;

	memset(&replace, 0, sizeof(replace));

	for (i = 0, len = 0; i < repl->num; i++) 
		len += repl->len[i];

	replace.strr_pid = pid;
	replace.strr_nrepl = repl->num;
	replace.strr_base = malloc(len);
	replace.strr_len = len;
	if (replace.strr_base == NULL)
		err(1, "%s: malloc", __func__);
	for (i = 0, off = 0; i < repl->num; i++) {
		replace.strr_argind[i] = repl->ind[i];
		replace.strr_offlen[i] = repl->len[i];
		if (repl->len[i] == 0) {
			replace.strr_off[i] = (size_t)repl->address[i];
			continue;
		}

		replace.strr_off[i] = off;
		memcpy(replace.strr_base + off,
		    repl->address[i], repl->len[i]);

		off += repl->len[i];
	}

	if ((ret = ioctl(fd, STRIOCREPLACE, &replace)) == -1)
		warn("%s:%d %s: ioctl", __FILE__, __LINE__, __func__);

	free(replace.strr_base);

	return (ret);
}

static int
linux_io(int fd, pid_t pid, int op, void *addr, u_char *buf, size_t size)
{
	struct systrace_io io;

	memset(&io, 0, sizeof(io));
	io.strio_pid = pid;
	io.strio_addr = buf;
	io.strio_len = size;
	io.strio_offs = addr;
	io.strio_op = (op == INTERCEPT_READ ? SYSTR_READ : SYSTR_WRITE);

	return (ioctl(fd, STRIOCIO, &io) == -1 ? -1 : 0);
}

static int
linux_setcwd(int fd, pid_t pid)
{
	return (ioctl(fd, STRIOCGETCWD, &pid));
}

static int
linux_restcwd(int fd)
{
	int res;

	if ((res = ioctl(fd, STRIOCRESCWD, 0)) == -1)
		warn("%s:%d %s: ioctl", __FILE__, __LINE__, __func__); /* XXX */

	return (res);
}

static int
linux_argument(int off, void *pargs, int argsize, void **pres)
{
	register_t *args = (register_t *)pargs;
	
	if (off >= argsize / sizeof(register_t))
		return (-1);

	*pres = (void *)args[off];

	return (0);
}

static int
linux_read(int fd)
{
	struct str_message msg;
	struct intercept_pid *icpid;

	const char *sysname;
	u_int16_t seqnr;
	pid_t pid;
	int code;

	if (read(fd, &msg, sizeof(msg)) != sizeof(msg))
		return (-1);

	icpid = linux_getpid(msg.msg_pid);
	if (icpid == NULL)
		return (-1);

	seqnr = msg.msg_seqnr;
	pid = msg.msg_pid;
	switch (msg.msg_type) {
	case SYSTR_MSG_ASK:
		code = msg.msg_data.msg_ask.code;
		sysname = linux_syscall_name(pid, code);

		intercept_syscall(fd, pid, seqnr, msg.msg_policy,

		    sysname, code, "linux",
		    (void *)msg.msg_data.msg_ask.args,
		    msg.msg_data.msg_ask.argsize);
		break;

	case SYSTR_MSG_RES:
		code = msg.msg_data.msg_ask.code;
		sysname = linux_syscall_name(pid, code);

		intercept_syscall_result(fd, pid, seqnr, msg.msg_policy,
		    sysname, code, "linux",
		    (void *)msg.msg_data.msg_ask.args,
		    msg.msg_data.msg_ask.argsize,
		    msg.msg_data.msg_ask.result,
		    msg.msg_data.msg_ask.rval);
		break;

	case SYSTR_MSG_UGID: {
		struct str_msg_ugid *msg_ugid = &msg.msg_data.msg_ugid;

		intercept_ugid(icpid, msg_ugid->uid, msg_ugid->uid);

		if (linux_answer(fd, pid, seqnr, 0, 0, 0, NULL) == -1)
			err(1, "%s:%d: answer", __func__, __LINE__);
		break;
	}
	case SYSTR_MSG_CHILD:
		intercept_child_info(msg.msg_pid, msg.msg_data.msg_child.new_pid);
		break;

#ifdef SYSTR_MSG_POLICYFREE
	case SYSTR_MSG_POLICYFREE:
		intercept_policy_free(msg.msg_policy);
		break;
#endif
	}
	return (0);
}


struct intercept_system intercept = {
	"linux kernel",
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
