/*	$OpenBSD: openbsd-syscalls.c,v 1.10 2002/07/30 09:16:19 itojun Exp $	*/
/*
 * Copyright 2002 Niels Provos <provos@citi.umich.edu>
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
 *      This product includes software developed by Niels Provos.
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
#include <sys/param.h>

#include <sys/syscall.h>

#include <sys/ioctl.h>
#include <sys/tree.h>
#include <sys/systrace.h>

#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>
#include <err.h>

#include "intercept.h"

char *syscallnames[] = {
        "syscall",		/* 0 = syscall */
        "exit",		/* 1 = exit */
        "fork",		/* 2 = fork */
        "read",		/* 3 = read */
        "write",		/* 4 = write */
        "open",		/* 5 = open */
        "close",		/* 6 = close */
        "wait4",		/* 7 = wait4 */
        "obs_creat",		/* 8 = old creat */
        "link",		/* 9 = link */
        "unlink",		/* 10 = unlink */
        "obs_execv",		/* 11 = obsolete execv */
        "chdir",		/* 12 = chdir */
        "fchdir",		/* 13 = fchdir */
        "mknod",		/* 14 = mknod */
        "chmod",		/* 15 = chmod */
        "chown",		/* 16 = chown */
        "obs_break",		/* 17 = obsolete break */
        "old_obs_getfsstat",		/* 18 = obsolete getfsstat */
        "old_lseek",		/* 19 = old lseek */
        "getpid",		/* 20 = getpid */
        "obs_mount",		/* 21 = obsolete mount */
        "obs_unmount",		/* 22 = obsolete unmount */
        "setuid",		/* 23 = setuid */
        "getuid",		/* 24 = getuid */
        "geteuid",		/* 25 = geteuid */
        "ptrace",		/* 26 = ptrace */
        "recvmsg",		/* 27 = recvmsg */
        "sendmsg",		/* 28 = sendmsg */
        "recvfrom",		/* 29 = recvfrom */
        "accept",		/* 30 = accept */
        "getpeername",		/* 31 = getpeername */
        "getsockname",		/* 32 = getsockname */
        "access",		/* 33 = access */
        "chflags",		/* 34 = chflags */
        "fchflags",		/* 35 = fchflags */
        "sync",		/* 36 = sync */
        "kill",		/* 37 = kill */
        "obs_stat",		/* 38 = old stat */
        "getppid",		/* 39 = getppid */
        "obs_lstat",		/* 40 = old lstat */
        "dup",		/* 41 = dup */
        "pipe",		/* 42 = pipe */
        "getegid",		/* 43 = getegid */
        "profil",		/* 44 = profil */
        "ktrace",		/* 45 = ktrace */
        "sigaction",		/* 46 = sigaction */
        "getgid",		/* 47 = getgid */
        "sigprocmask",		/* 48 = sigprocmask */
        "getlogin",		/* 49 = getlogin */
        "setlogin",		/* 50 = setlogin */
        "acct",		/* 51 = acct */
        "sigpending",		/* 52 = sigpending */
        "sigaltstack",		/* 53 = sigaltstack */
        "ioctl",		/* 54 = ioctl */
        "reboot",		/* 55 = reboot */
        "revoke",		/* 56 = revoke */
        "symlink",		/* 57 = symlink */
        "readlink",		/* 58 = readlink */
        "execve",		/* 59 = execve */
        "umask",		/* 60 = umask */
        "chroot",		/* 61 = chroot */
        "obs_fstat",		/* 62 = old fstat */
        "#63",		/* 63 = reserved */
        "obs_getpagesize",		/* 64 = old getpagesize */
        "msync",		/* 65 = msync */
        "vfork",		/* 66 = vfork */
        "obs_vread",		/* 67 = obsolete vread */
        "obs_vwrite",		/* 68 = obsolete vwrite */
        "sbrk",		/* 69 = sbrk */
        "sstk",		/* 70 = sstk */
        "obs_mmap",		/* 71 = old mmap */
        "obs_vadvise",		/* 72 = obsolete vadvise */
        "munmap",		/* 73 = munmap */
        "mprotect",		/* 74 = mprotect */
        "madvise",		/* 75 = madvise */
        "#76",		/* 76 = obsolete vhangup */
        "#77",		/* 77 = obsolete vlimit */
        "mincore",		/* 78 = mincore */
        "getgroups",		/* 79 = getgroups */
        "setgroups",		/* 80 = setgroups */
        "getpgrp",		/* 81 = getpgrp */
        "setpgid",		/* 82 = setpgid */
        "setitimer",		/* 83 = setitimer */
        "old_wait",		/* 84 = old wait */
        "obs_swapon",		/* 85 = swapon */
        "getitimer",		/* 86 = getitimer */
        "obs_gethostname",		/* 87 = old gethostname */
        "obs_sethostname",		/* 88 = old sethostname */
        "getdtablesize",		/* 89 = getdtablesize */
        "dup2",		/* 90 = dup2 */
        "#91",		/* 91 = getdopt */
        "fcntl",		/* 92 = fcntl */
        "select",		/* 93 = select */
        "#94",		/* 94 = setdopt */
        "fsync",		/* 95 = fsync */
        "setpriority",		/* 96 = setpriority */
        "socket",		/* 97 = socket */
        "connect",		/* 98 = connect */
        "obs_accept",		/* 99 = old accept */
        "getpriority",		/* 100 = getpriority */
        "old_send",		/* 101 = old send */
        "old_recv",		/* 102 = old recv */
        "sigreturn",		/* 103 = sigreturn */
        "bind",		/* 104 = bind */
        "setsockopt",		/* 105 = setsockopt */
        "listen",		/* 106 = listen */
        "#107",		/* 107 = obsolete vtimes */
        "obs_sigvec",		/* 108 = old sigvec */
        "obs_sigblock",		/* 109 = old sigblock */
        "obs_sigsetmask",		/* 110 = old sigsetmask */
        "sigsuspend",		/* 111 = sigsuspend */
        "obs_sigstack",		/* 112 = old sigstack */
        "obs_recvmsg",		/* 113 = old recvmsg */
        "obs_sendmsg",		/* 114 = old sendmsg */
        "#115",		/* 115 = obsolete vtrace */
        "gettimeofday",		/* 116 = gettimeofday */
        "getrusage",		/* 117 = getrusage */
        "getsockopt",		/* 118 = getsockopt */
        "#119",		/* 119 = nosys */
        "readv",		/* 120 = readv */
        "writev",		/* 121 = writev */
        "settimeofday",		/* 122 = settimeofday */
        "fchown",		/* 123 = fchown */
        "fchmod",		/* 124 = fchmod */
        "obs_recvfrom",		/* 125 = old recvfrom */
        "obs_setreuid",		/* 126 = old setreuid */
        "obs_setregid",		/* 127 = old setregid */
        "rename",		/* 128 = rename */
        "obs_truncate",		/* 129 = old truncate */
        "obs_ftruncate",		/* 130 = old ftruncate */
        "flock",		/* 131 = flock */
        "mkfifo",		/* 132 = mkfifo */
        "sendto",		/* 133 = sendto */
        "shutdown",		/* 134 = shutdown */
        "socketpair",		/* 135 = socketpair */
        "mkdir",		/* 136 = mkdir */
        "rmdir",		/* 137 = rmdir */
        "utimes",		/* 138 = utimes */
        "futimes",		/* 139 = futimes */
        "adjtime",		/* 140 = adjtime */
        "obs_getpeername",		/* 141 = old getpeername */
        "obs_gethostid",		/* 142 = old gethostid */
        "#143",		/* 143 = old sethostid */
        "obs_getrlimit",		/* 144 = old getrlimit */
        "obs_setrlimit",		/* 145 = old setrlimit */
        "obs_killpg",		/* 146 = old killpg */
        "setsid",		/* 147 = setsid */
        "#148",		/* 148 = obsolete setquota */
        "#149",		/* 149 = obsolete qquota */
        "obs_getsockname",		/* 150 = old getsockname */
        "#151",		/* 151 = nosys */
        "setprivexec",		/* 152 = setprivexec */
        "pread",		/* 153 = pread */
        "pwrite",		/* 154 = pwrite */
        "nfssvc",		/* 155 = nfssvc */
        "old_getdirentries",		/* 156 =getdirentries */
        "statfs",		/* 157 = statfs */
        "fstatfs",		/* 158 = fstatfs */
        "unmount",		/* 159 = unmount */
        "#160",		/* 160 = obsolete async_daemon */
        "getfh",		/* 161 = getfh */
        "obs_getdomainname",		/* 162 = old getdomainname */
        "obs_setdomainname",		/* 163 = old setdomainname */
        "#164",		/* 164 */
        "quotactl",		/* 165 = quotactl */
        "#166",		/* 166 = obsolete exportfs */
        "mount",		/* 167 = mount */
        "#168",		/* 168 = obsolete ustat */
        "#169",		/* 169 = nosys */
        "#170",		/* 170 = obsolete table */
        "obs_wait3",		/* 171 = old wait3 */
        "#172",		/* 172 = obsolete rpause */
        "#173",		/* 173 = nosys */
        "#174",		/* 174 = obsolete getdents */
        "#175",		/* 175 = nosys */
        "add_profil",		/* NeXT */
        "#177",		/* 177 = nosys */
        "#178",		/* 178 = nosys */
        "#179",		/* 179 = nosys */
        "kdebug_trace",		/* 180 = kdebug_trace */
        "setgid",		/* 181 = setgid */
        "setegid",		/* 182 = setegid */
        "seteuid",		/* 183 = seteuid */
        "#184",		/* 184 = nosys */
        "#185",		/* 185 = nosys */
        "#186",		/* 186 = nosys */
        "#187",		/* 187 = nosys */
        "stat",		/* 188 = stat */
        "fstat",		/* 189 = fstat */
        "lstat",		/* 190 = lstat */
        "pathconf",		/* 191 = pathconf */
        "fpathconf",		/* 192 = fpathconf */
        "obs_getfsstat",		/* 193 = old getfsstat */
        "getrlimit",		/* 194 = getrlimit */
        "setrlimit",		/* 195 = setrlimit */
        "getdirentries",		/* 196 = getdirentries */
        "mmap",		/* 197 = mmap */
        "#198",		/* 198 = __syscall */
        "lseek",		/* 199 = lseek */
        "truncate",		/* 200 = truncate */
        "ftruncate",		/* 201 = ftruncate */
        "__sysctl",		/* 202 = __sysctl */
        "mlock",		/* 203 = mlock */
        "munlock",		/* 204 = munlock */
        "undelete",		/* 205 = undelete */
        "ATsocket",		/* 206 = ATsocket */
        "ATgetmsg",		/* 207 = ATgetmsg */
        "ATputmsg",		/* 208 = ATputmsg */
        "ATPsndreq",		/* 209 = ATPsndreq */
        "ATPsndrsp",		/* 210 = ATPsndrsp */
        "ATPgetreq",		/* 211 = ATPgetreq */
        "ATPgetrsp",		/* 212 = ATPgetrsp */
        "#213",		/* 213 = Reserved for AppleTalk */
        "#214",		/* 214 = Reserved for AppleTalk */
        "#215",		/* 215 = Reserved for AppleTalk */
        "#216",		/* 216 = Reserved */
        "#217",		/* 217 = Reserved */
        "#218",		/* 218 = Reserved */
        "#219",		/* 219 = Reserved */
        "getattrlist",		/* 220 = getattrlist */
        "setattrlist",		/* 221 = setattrlist */
        "getdirentriesattr",		/* 222 = getdirentriesattr */
        "exchangedata",		/* 223 = exchangedata */
        "checkuseraccess",		/* 224 - checkuseraccess */
        "searchfs",		/* 225 = searchfs */
        "delete",		/* 226 = private delete call */
        "copyfile",		/* 227 = copyfile  */
        "#228",		/* 228 = nosys */
        "#229",		/* 229 = nosys */
        "#230",		/* 230 = reserved for AFS */
        "watchevent",		/* 231 = watchevent */
        "waitevent",		/* 232 = waitevent */
        "modwatch",		/* 233 = modwatch */
        "#234",		/* 234 = nosys */
        "#235",		/* 235 = nosys */
        "#236",		/* 236 = nosys */
        "#237",		/* 237 = nosys */
        "#238",		/* 238 = nosys */
        "#239",		/* 239 = nosys */
        "#240",		/* 240 = nosys */
        "#241",		/* 241 = nosys */
        "fsctl",		/* 242 = fsctl */
        "#243",		/* 243 = nosys */
        "#244",		/* 244 = nosys */
        "#245",		/* 245 = nosys */
        "#246",		/* 246 = nosys */
        "#247",		/* 247 = nosys */
        "#248",		/* 248 = nosys */
        "#249",		/* 249 = nosys */
        "minherit",		/* 250 = minherit */
        "semsys",		/* 251 = semsys */
        "msgsys",		/* 252 = msgsys */
        "shmsys",		/* 253 = shmsys */
        "semctl",		/* 254 = semctl */
        "semget",		/* 255 = semget */
        "semop",		/* 256 = semop */
        "semconfig",		/* 257 = semconfig */
        "msgctl",		/* 258 = msgctl */
        "msgget",		/* 259 = msgget */
        "msgsnd",		/* 260 = msgsnd */
        "msgrcv",		/* 261 = msgrcv */
        "shmat",		/* 262 = shmat */
        "shmctl",		/* 263 = shmctl */
        "shmdt",		/* 264 = shmdt */
        "shmget",		/* 265 = shmget */
        "shm_open",		/* 266 = shm_open */
        "shm_unlink",		/* 267 = shm_unlink */
        "sem_open",		/* 268 = sem_open */
        "sem_close",		/* 269 = sem_close */
        "sem_unlink",		/* 270 = sem_unlink */
        "sem_wait",		/* 271 = sem_wait */
        "sem_trywait",		/* 272 = sem_trywait */
        "sem_post",		/* 273 = sem_post */
        "sem_getvalue",		/* 274 = sem_getvalue */
        "sem_init",		/* 275 = sem_init */
        "sem_destroy",		/* 276 = sem_destroy */
        "#277",		/* 277 = nosys */
        "#278",		/* 278 = nosys */
        "#279",		/* 279 = nosys */
        "#280",		/* 280 = nosys */
        "#281",		/* 281 = nosys */
        "#282",		/* 282 = nosys */
        "#283",		/* 283 = nosys */
        "#284",		/* 284 = nosys */
        "#285",		/* 285 = nosys */
        "#286",		/* 286 = nosys */
        "#287",		/* 287 = nosys */
        "#288",		/* 288 = nosys */
        "#289",		/* 289 = nosys */
        "#290",		/* 290 = nosys */
        "#291",		/* 291 = nosys */
        "#292",		/* 292 = nosys */
        "#293",		/* 293 = nosys */
        "#294",		/* 294 = nosys */
        "#295",		/* 295 = nosys */
        "load_shared_file",		/* 296 = load_shared_file */
        "reset_shared_file",		/* 297 = reset_shared_file */
        "new_system_shared_regions",		/* 298 = new_system_shared_regions */
        "#299",		/* 299 = nosys */
        "#300",		/* 300 = modnext */
        "#301",		/* 301 = modstat */
        "#302",		/* 302 = modfnext */
        "#303",		/* 303 = modfind */
        "#304",		/* 304 = kldload */
        "#305",		/* 305 = kldunload */
        "#306",		/* 306 = kldfind */
        "#307",		/* 307 = kldnext */
        "#308",		/* 308 = kldstat */
        "#309",		/* 309 = kldfirstmod */
        "#310",		/* 310 = getsid */
        "#311",		/* 311 = setresuid */
        "#312",		/* 312 = setresgid */
        "#313",		/* 313 = obsolete signanosleep */
        "#314",		/* 314 = aio_return */
        "#315",		/* 315 = aio_suspend */
        "#316",		/* 316 = aio_cancel */
        "#317",		/* 317 = aio_error */
        "#318",		/* 318 = aio_read */
        "#319",		/* 319 = aio_write */
        "#320",		/* 320 = lio_listio */
        "#321",		/* 321 = yield */
        "#322",		/* 322 = thr_sleep */
        "#323",		/* 323 = thr_wakeup */
        "mlockall",		/* 324 = mlockall */
        "munlockall",		/* 325 = munlockall */
        "#326",		/* 326 */
        "issetugid",		/* 327 = issetugid */
        "__pthread_kill",		/* 328  = __pthread_kill */
        "pthread_sigmask",		/* 329  = pthread_sigmask */
        "sigwait",		/* 330 = sigwait */
        "#331",		/* 331 */
        "#332",		/* 332 */
        "#333",		/* 333 */
        "#334",		/* 334 */
        "utrace",		/* 335 = utrace */
        "#336",		/* 336 */
        "#337",		/* 337 */
        "#338",		/* 338 */
        "#339",		/* 339 */
        "#340",		/* 340 */
        "#341",		/* 341 */
        "#342",		/* 342 */
        "#343",		/* 343 */
        "#344",		/* 344 */
        "#345",		/* 345 */
        "#346",		/* 346 */
        "#347",		/* 347 */
        "#348",		/* 348 */
        "#349"		/* 349 */
};

struct emulation {
	const char *name;	/* Emulation name */
	char **sysnames;	/* Array of system call names */
	int  nsysnames;		/* Number of */
};

#define SYS_MAXSYSCALL	349

static struct emulation emulations[] = {
	{ "darwin",	syscallnames,		SYS_MAXSYSCALL },
	{ NULL,		NULL,			NULL }
};

struct darwin_data {
	struct emulation *current;
	struct emulation *commit;
};

static int darwin_init(void);
static int darwin_attach(int, pid_t);
static int darwin_report(int, pid_t);
static int darwin_detach(int, pid_t);
static int darwin_open(void);
static struct intercept_pid *darwin_getpid(pid_t);
static void darwin_freepid(struct intercept_pid *);
static void darwin_clonepid(struct intercept_pid *, struct intercept_pid *);
static struct emulation *darwin_find_emulation(const char *);
static int darwin_set_emulation(pid_t, const char *);
static struct emulation *darwin_switch_emulation(struct darwin_data *);
static const char *darwin_syscall_name(pid_t, int);
static int darwin_syscall_number(const char *, const char *);
static short darwin_translate_policy(short);
static short darwin_translate_flags(short);
static int darwin_translate_errno(int);
static int darwin_answer(int, pid_t, u_int32_t, short, int, short,
    struct elevate *);
static int darwin_newpolicy(int);
static int darwin_assignpolicy(int, pid_t, int);
static int darwin_modifypolicy(int, int, int, short);
static int darwin_replace(int, pid_t, struct intercept_replace *);
static int darwin_io(int, pid_t, int, void *, u_char *, size_t);
static int darwin_setcwd(int, pid_t);
static int darwin_restcwd(int);
static int darwin_argument(int, void *, int, void **);
static int darwin_read(int);

static int
darwin_init(void)
{
	return (0);
}

static int
darwin_attach(int fd, pid_t pid)
{
	if (ioctl(fd, STRIOCATTACH, &pid) == -1)
		return (-1);

	return (0);
}

static int
darwin_report(int fd, pid_t pid)
{
	if (ioctl(fd, STRIOCREPORT, &pid) == -1)
		return (-1);

	return (0);
}

static int
darwin_detach(int fd, pid_t pid)
{
	if (ioctl(fd, STRIOCDETACH, &pid) == -1)
		return (-1);

	return (0);
}

static int
darwin_open(void)
{
	char *path = "/dev/systrace";
	int fd, cfd = -1;

	fd = open(path, O_RDONLY, 0);
	if (fd == -1) {
		warn("open: %s", path);
		return (-1);
	}

	if (ioctl(fd, SYSTR_CLONE, &cfd) == -1) {
		warn("ioctl(SYSTR_CLONE)");
		goto out;
	}

	if (fcntl(cfd, F_SETFD, 1) == -1)
		warn("fcntl(F_SETFD)");

 out:
	close (fd);
	return (cfd);
}

static struct intercept_pid *
darwin_getpid(pid_t pid)
{
	struct intercept_pid *icpid;
	struct darwin_data *data;

	icpid = intercept_getpid(pid);
	if (icpid == NULL)
		return (NULL);
	if (icpid->data != NULL)
		return (icpid);

	if ((icpid->data = malloc(sizeof(struct darwin_data))) == NULL)
		err(1, "%s:%d: malloc", __func__, __LINE__);

	data = icpid->data;
	data->current = &emulations[0];
	data->commit = NULL;

	return (icpid);
}

static void
darwin_freepid(struct intercept_pid *ipid)
{
	if (ipid->data != NULL)
		free(ipid->data);
}

static void
darwin_clonepid(struct intercept_pid *opid, struct intercept_pid *npid)
{
	if (opid->data == NULL) {
		npid->data = NULL;
		return;
	}

	if ((npid->data = malloc(sizeof(struct darwin_data))) == NULL)
		err(1, "%s:%d: malloc", __func__, __LINE__);
	memcpy(npid->data, opid->data, sizeof(struct darwin_data));
}

static struct emulation *
darwin_find_emulation(const char *name)
{
	struct emulation *tmp;

	tmp = emulations;
	while (tmp->name) {
		if (!strcmp(tmp->name, name))
			break;
		tmp++;
	}

	if (!tmp->name)
		return (NULL);

	return (tmp);
}

static int
darwin_set_emulation(pid_t pidnr, const char *name)
{
	struct emulation *tmp;
	struct intercept_pid *pid;
	struct darwin_data *data;

	if ((tmp = darwin_find_emulation(name)) == NULL)
		return (-1);

	pid = intercept_getpid(pidnr);
	if (pid == NULL)
		return (-1);
	data = pid->data;

	data->commit = tmp;

	return (0);
}

static struct emulation *
darwin_switch_emulation(struct darwin_data *data)
{
	data->current = data->commit;
	data->commit = NULL;

	return (data->current);
}

static const char *
darwin_syscall_name(pid_t pidnr, int number)
{
	struct intercept_pid *pid;
	struct emulation *current;

	pid = darwin_getpid(pidnr);
	if (pid == NULL)
		return (NULL);
	current = ((struct darwin_data *)pid->data)->current;

	if (number < 0 || number >= current->nsysnames)
		return (NULL);

	return (current->sysnames[number]);
}

static int
darwin_syscall_number(const char *emulation, const char *name)
{
	struct emulation *current;
	int i;

	current = darwin_find_emulation(emulation);
	if (current == NULL)
		return (-1);

	for (i = 0; i < current->nsysnames; i++)
		if (!strcmp(name, current->sysnames[i]))
			return (i);

	return (-1);
}

static short
darwin_translate_policy(short policy)
{
	switch (policy) {
	case ICPOLICY_ASK:
		return (SYSTR_POLICY_ASK);
	case ICPOLICY_PERMIT:
		return (SYSTR_POLICY_PERMIT);
	case ICPOLICY_NEVER:
	default:
		return (SYSTR_POLICY_NEVER);
	}
}

static short
darwin_translate_flags(short flags)
{
	switch (flags) {
	case ICFLAGS_RESULT:
		return (SYSTR_FLAGS_RESULT);
	default:
		return (0);
	}
}

static int
darwin_translate_errno(int nerrno)
{
	return (nerrno);
}

static int
darwin_answer(int fd, pid_t pid, u_int32_t seqnr, short policy, int nerrno,
    short flags, struct elevate *elevate)
{
	struct systrace_answer ans;

	memset(&ans, 0, sizeof(ans));
	ans.stra_pid = pid;
	ans.stra_seqnr = seqnr;
	ans.stra_policy = darwin_translate_policy(policy);
	ans.stra_flags = darwin_translate_flags(flags);
	ans.stra_error = darwin_translate_errno(nerrno);

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

	if (ioctl(fd, STRIOCANSWER, &ans) == -1) {
		warn("%s:%d ioctl", __func__, __LINE__);
		return (-1);
	}

	return (0);
}

static int
darwin_newpolicy(int fd)
{
	struct systrace_policy pol;

	pol.strp_op = SYSTR_POLICY_NEW;
	pol.strp_num = -1;
	pol.strp_maxents = 512;

	if (ioctl(fd, STRIOCPOLICY, &pol) == -1)
		return (-1);

	return (pol.strp_num);
}

static int
darwin_assignpolicy(int fd, pid_t pid, int num)
{
	struct systrace_policy pol;

	pol.strp_op = SYSTR_POLICY_ASSIGN;
	pol.strp_num = num;
	pol.strp_pid = pid;

	if (ioctl(fd, STRIOCPOLICY, &pol) == -1)
		return (-1);

	return (0);
}

static int
darwin_modifypolicy(int fd, int num, int code, short policy)
{
	struct systrace_policy pol;

	pol.strp_op = SYSTR_POLICY_MODIFY;
	pol.strp_num = num;
	pol.strp_code = code;
	pol.strp_policy = darwin_translate_policy(policy);

	if (ioctl(fd, STRIOCPOLICY, &pol) == -1)
		return (-1);

	return (0);
}

static int
darwin_replace(int fd, pid_t pid, struct intercept_replace *repl)
{
	struct systrace_replace replace;
	size_t len, off;
	int i, ret;

	memset(&replace, 0, sizeof(replace));

	for (i = 0, len = 0; i < repl->num; i++) {
		len += repl->len[i];
	}

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

	ret = ioctl(fd, STRIOCREPLACE, &replace);
	if (ret == -1 && errno != EBUSY) {
		warn("%s: ioctl", __func__);
	}

	free(replace.strr_base);
	
	return (ret);
}

static int
darwin_io(int fd, pid_t pid, int op, void *addr, u_char *buf, size_t size)
{
	struct systrace_io io;
	extern int ic_abort;

	memset(&io, 0, sizeof(io));
	io.strio_pid = pid;
	io.strio_addr = buf;
	io.strio_len = size;
	io.strio_offs = addr;
	io.strio_op = (op == INTERCEPT_READ ? SYSTR_READ : SYSTR_WRITE);
	if (ioctl(fd, STRIOCIO, &io) == -1) {
		if (errno == EBUSY)
			ic_abort = 1;
		return (-1);
	}

	return (0);
}

static int
darwin_setcwd(int fd, pid_t pid)
{
	return (ioctl(fd, STRIOCGETCWD, &pid));
}

static int
darwin_restcwd(int fd)
{
	int res;
	if ((res = ioctl(fd, STRIOCRESCWD, 0)) == -1)
		warn("%s: ioctl", __func__); /* XXX */

	return (res);
}

static int
darwin_argument(int off, void *pargs, int argsize, void **pres)
{
	register_t *args = (register_t *)pargs;

	if (off >= argsize / sizeof(register_t))
		return (-1);

	*pres = (void *)args[off];

	return (0);
}

static int
darwin_read(int fd)
{
	struct str_message msg;
	struct intercept_pid *icpid;
	struct darwin_data *data;
	struct emulation *current;

	char name[SYSTR_EMULEN+1];
	const char *sysname;
	u_int16_t seqnr;
	pid_t pid;
	int code;

	if (read(fd, &msg, sizeof(msg)) != sizeof(msg))
		return (-1);

	icpid = darwin_getpid(msg.msg_pid);
	if (icpid == NULL)
		return (-1);
	data = icpid->data;

	current = data->current;
	
	seqnr = msg.msg_seqnr;
	pid = msg.msg_pid;
	switch (msg.msg_type) {
	case SYSTR_MSG_ASK:
		code = msg.msg_data.msg_ask.code;
		sysname = darwin_syscall_name(pid, code);

		intercept_syscall(fd, pid, seqnr, msg.msg_policy,
		    sysname, code, current->name,
		    (void *)msg.msg_data.msg_ask.args,
		    msg.msg_data.msg_ask.argsize);
		break;

	case SYSTR_MSG_RES:
		code = msg.msg_data.msg_ask.code;
		sysname = darwin_syscall_name(pid, code);

		/* Switch emulation around at the right time */
		if (data->commit != NULL) {
			current = darwin_switch_emulation(data);
		}

		intercept_syscall_result(fd, pid, seqnr, msg.msg_policy,
		    sysname, code, current->name,
		    (void *)msg.msg_data.msg_ask.args,
		    msg.msg_data.msg_ask.argsize,
		    msg.msg_data.msg_ask.result,
		    msg.msg_data.msg_ask.rval);
		break;

	case SYSTR_MSG_EMUL:
		memcpy(name, msg.msg_data.msg_emul.emul, SYSTR_EMULEN);
		name[SYSTR_EMULEN] = '\0';

		if (darwin_set_emulation(pid, name) == -1)
			errx(1, "%s:%d: set_emulation(%s)",
			    __func__, __LINE__, name);

		if (icpid->execve_code == -1) {
			icpid->execve_code = 0;

			/* A running attach fake a exec cb */
			current = darwin_switch_emulation(data);

			intercept_syscall_result(fd,
			    pid, seqnr, msg.msg_policy,
			    "execve", 0, current->name,
			    NULL, 0, 0, NULL);
			break;
		}

		if (darwin_answer(fd, pid, seqnr, 0, 0, 0, NULL) == -1)
			err(1, "%s:%d: answer", __func__, __LINE__);
		break;

	case SYSTR_MSG_UGID: {
		struct str_msg_ugid *msg_ugid;
		
		msg_ugid = &msg.msg_data.msg_ugid;

		intercept_ugid(icpid, msg_ugid->uid, msg_ugid->uid);

		if (darwin_answer(fd, pid, seqnr, 0, 0, 0, NULL) == -1)
			err(1, "%s:%d: answer", __func__, __LINE__);
		break;
	}
	case SYSTR_MSG_CHILD:
		intercept_child_info(msg.msg_pid,
		    msg.msg_data.msg_child.new_pid);
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
	"darwin",
	darwin_init,
	darwin_open,
	darwin_attach,
	darwin_detach,
	darwin_report,
	darwin_read,
	darwin_syscall_number,
	darwin_setcwd,
	darwin_restcwd,
	darwin_io,
	darwin_argument,
	darwin_answer,
	darwin_newpolicy,
	darwin_assignpolicy,
	darwin_modifypolicy,
	darwin_replace,
	darwin_clonepid,
	darwin_freepid,
};
