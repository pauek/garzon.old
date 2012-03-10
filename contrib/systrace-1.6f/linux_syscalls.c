/*
 * Copyright (c) 2002 Marius Aamodt Eriksen <marius@umich.edu>
 */

#ifndef NR_syscalls
#define NR_syscalls MAX_SYSCALLS
#endif

#include <assert.h>

/*
 * this file emulates openbsd's syscallnames[] array.  linux does not
 * have such a feature.  automatically generated from
 * arch/i386/kernel/entry.S
 */

char *linux_syscallnames[] = {
	"ni_syscall-1",
	"exit",
	"fork",
	"read",
	"write",
	"open",             /* 5 */
	"close",
	"waitpid",
	"creat",
	"link",
	"unlink",           /* 10 */
	"execve",
	"chdir",
	"time",
	"mknod",
	"chmod",            /* 15 */
	"lchown16",
	"ni_syscall-2",                       /* old break syscall holder */
	"stat",
	"lseek",
	"getpid",           /* 20 */
	"mount",
	"oldumount",
	"setuid16",
	"getuid16",
	"stime",            /* 25 */
	"ptrace",
	"alarm",
	"fstat",
	"pause",
	"utime",            /* 30 */
	"ni_syscall-3",       /* old stty syscall holder */
	"ni_syscall-4",       /* old gtty syscall holder */
	"access",
	"nice",
	"ni_syscall-5",       /* 35 */   /* old ftime syscall holder */
	"sync",
	"kill",
	"rename",
	"mkdir",
	"rmdir",            /* 40 */
	"dup",
	"pipe",
	"times",
	"ni_syscall-6",       /* old prof syscall holder */
	"brk",              /* 45 */
	"setgid16",
	"getgid16",
	"signal",
	"geteuid16",
	"getegid16",        /* 50 */
	"acct",
	"umount",           /* recycled never used phys() */
	"ni_syscall-7",       /* old lock syscall holder */
	"ioctl",
	"fcntl",            /* 55 */
	"ni_syscall-8",       /* old mpx syscall holder */
	"setpgid",
	"ni_syscall-9",       /* old ulimit syscall holder */
	"olduname",
	"umask",            /* 60 */
	"chroot",
	"ustat",
	"dup2",
	"getppid",
	"getpgrp",          /* 65 */
	"setsid",
	"sigaction",
	"sgetmask",
	"ssetmask",
	"setreuid16",       /* 70 */
	"setregid16",
	"sigsuspend",
	"sigpending",
	"sethostname",
	"setrlimit",        /* 75 */
	"old_getrlimit",
	"getrusage",
	"gettimeofday",
	"settimeofday",
	"getgroups16",      /* 80 */
	"setgroups16",
	"old_select",
	"symlink",
	"lstat",
	"readlink",         /* 85 */
	"uselib",
	"swapon",
	"reboot",
	"old_readdir",
	"old_mmap",             /* 90 */
	"munmap",
	"truncate",
	"ftruncate",
	"fchmod",
	"fchown16",         /* 95 */
	"getpriority",
	"setpriority",
	"ni_syscall-10",       /* old profil syscall holder */
	"statfs",
	"fstatfs",          /* 100 */
	"ioperm",
	"socketcall",
	"syslog",
	"setitimer",
	"getitimer",        /* 105 */
	"newstat",
	"newlstat",
	"newfstat",
	"uname",
	"iopl",             /* 110 */
	"vhangup",
	"ni_syscall-11",       /* old "idle" system call */
	"vm86old",
	"wait4",
	"swapoff",          /* 115 */
	"sysinfo",
	"ipc",
	"fsync",
	"sigreturn",
	"clone",            /* 120 */
	"setdomainname",
	"newuname",
	"modify_ldt",
	"adjtimex",
	"mprotect",         /* 125 */
	"sigprocmask",
	"create_module",
	"init_module",
	"delete_module",
	"get_kernel_syms",  /* 130 */
	"quotactl",
	"getpgid",
	"fchdir",
	"bdflush",
	"sysfs",            /* 135 */
	"personality",
	"ni_syscall-12",       /* for afs_syscall */
	"setfsuid16",
	"setfsgid16",
	"llseek",           /* 140 */
	"getdents",
	"select",
	"flock",
	"msync",
	"readv",            /* 145 */
	"writev",
	"getsid",
	"fdatasync",
	"sysctl",
	"mlock",            /* 150 */
	"munlock",
	"mlockall",
	"munlockall",
	"sched_setparam",
	"sched_getparam",   /* 155 */
	"sched_setscheduler",
	"sched_getscheduler",
	"sched_yield",
	"sched_get_priority_max",
	"sched_get_priority_min",  /* 160 */
	"sched_rr_get_interval",
	"nanosleep",
	"mremap",
	"setresuid16",
	"getresuid16",      /* 165 */
	"vm86",
	"query_module",
	"poll",
	"nfsservctl",
	"setresgid16",      /* 170 */
	"getresgid16",
	"prctl",
	"rt_sigreturn",
	"rt_sigaction",
	"rt_sigprocmask",   /* 175 */
	"rt_sigpending",
	"rt_sigtimedwait",
	"rt_sigqueueinfo",
	"rt_sigsuspend",
	"pread",            /* 180 */
	"pwrite",
	"chown16",
	"getcwd",
	"capget",
	"capset",           /* 185 */
	"sigaltstack",
	"sendfile",
	"ni_syscall-13",       /* streams1 */
	"ni_syscall-14",       /* streams2 */
	"vfork",            /* 190 */
	"getrlimit",
	"mmap2",
	"truncate64",
	"ftruncate64",
	"stat64",           /* 195 */
	"lstat64",
	"fstat64",
	"lchown",
	"getuid",
	"getgid",           /* 200 */
	"geteuid",
	"getegid",
	"setreuid",
	"setregid",
	"getgroups",        /* 205 */
	"setgroups",
	"fchown",
	"setresuid",
	"getresuid",
	"setresgid",        /* 210 */
	"getresgid",
	"chown",
	"setuid",
	"setgid",
	"setfsuid",         /* 215 */
	"setfsgid",
	"pivot_root",
	"mincore",
	"madvise",
	"getdents64",       /* 220 */
	"fcntl64",
	"ni_syscall-15",       /* reserved for TUX */
	"ni_syscall-16",       /* Reserved for Security */
	"gettid",
	"readahead",        /* 225 */
	"setxattr",       /* reserved for setxattr */
	"lsetxattr",       /* reserved for lsetxattr */
	"fsetxattr",       /* reserved for fsetxattr */
	"getxattr",       /* reserved for getxattr */
	"lgetxattr",       /* 230 reserved for lgetxattr */
	"fgetxattr",       /* reserved for fgetxattr */
	"listxattr",       /* reserved for listxattr */
	"llistxattr",       /* reserved for llistxattr */
	"flistxattr",       /* reserved for flistxattr */
	"removexattr",       /* 235 reserved for removexattr */
	"lremovexattr",       /* reserved for lremovexattr */
	"fremovexattr",       /* reserved for fremovexattr */
	"tkill",
	"sendfile64",
        "futex",         /* 240 */
        "sched_setaffinity",
        "sched_getaffinity",
        "set_thread_area",
        "get_thread_area",
        "io_setup",      /* 245 */
        "io_destroy",
        "io_getevents",
        "io_submit",
        "io_cancel",
        "fadvise64",     /* 250 */
        "ni_syscall",
        "exit_group",
        "lookup_dcookie",
        "epoll_create",
        "epoll_ctl",     /* 255 */
        "epoll_wait",
        "remap_file_pages",
        "set_tid_address",
        "timer_create",
        "timer_settime",         /* 260 */
        "timer_gettime",
        "timer_getoverrun",
        "timer_delete",
        "clock_settime",
        "clock_gettime",         /* 265 */
        "clock_getres",
        "clock_nanosleep",
        "statfs64",
        "fstatfs64",     
        "tgkill",        /* 270 */
        "utimes",
        "fadvise64_64",
        "ni_syscall",    /* sys_vserver */
        "mbind",
        "get_mempolicy",
        "set_mempolicy",
        "mq_open",
        "mq_unlink",
        "mq_timedsend",
        "mq_timedreceive",       /* 280 */
        "mq_notify",
        "mq_getsetattr",
        "ni_syscall",            /* reserved for kexec */
        "waitid",
        "ni_syscall",            /* 285 */ /* available */
        "add_key",
        "request_key",
        "keyctl",
	"ni_syscall-31",
	"ni_syscall-32",	 /* 290 */
	"ni_syscall-33",
	"ni_syscall-34",
	"ni_syscall-35",
	"ni_syscall-36",
	"ni_syscall-37",	 /* 295 */
	"ni_syscall-38",
	"ni_syscall-39",
	"ni_syscall-40",
	"ni_syscall-41",
	NULL			 /* sentinel */
};

#ifdef PTRACE_LINUX64
/* automatically generated by the following script from Justin Cappos:
 *  grep define /usr/include/asm/unistd_64.h  | awk '{print $2, $3}' |	\
 *  awk '{if (NF == 2) { foo = "\"" substr($1,6) "\""; printf("%20s, /\* \
 *   %d *\/\n",foo, $2);}} END {printf("%20s\n","NULL");}'
 */

char *linux_syscallnames_64[] = {
	"read", /* 0 */
	"write", /* 1 */
	"open", /* 2 */
	"close", /* 3 */
	"stat", /* 4 */
	"fstat", /* 5 */
	"lstat", /* 6 */
	"poll", /* 7 */
	"lseek", /* 8 */
	"mmap", /* 9 */
	"mprotect", /* 10 */
	"munmap", /* 11 */
	"brk", /* 12 */
	"rt_sigaction", /* 13 */
	"rt_sigprocmask", /* 14 */
	"rt_sigreturn", /* 15 */
	"ioctl", /* 16 */
	"pread64", /* 17 */
	"pwrite64", /* 18 */
	"readv", /* 19 */
	"writev", /* 20 */
	"access", /* 21 */
	"pipe", /* 22 */
	"select", /* 23 */
	"sched_yield", /* 24 */
	"mremap", /* 25 */
	"msync", /* 26 */
	"mincore", /* 27 */
	"madvise", /* 28 */
	"shmget", /* 29 */
	"shmat", /* 30 */
	"shmctl", /* 31 */
	"dup", /* 32 */
	"dup2", /* 33 */
	"pause", /* 34 */
	"nanosleep", /* 35 */
	"getitimer", /* 36 */
	"alarm", /* 37 */
	"setitimer", /* 38 */
	"getpid", /* 39 */
	"sendfile", /* 40 */
	"socket", /* 41 */
	"connect", /* 42 */
	"accept", /* 43 */
	"sendto", /* 44 */
	"recvfrom", /* 45 */
	"sendmsg", /* 46 */
	"recvmsg", /* 47 */
	"shutdown", /* 48 */
	"bind", /* 49 */
	"listen", /* 50 */
	"getsockname", /* 51 */
	"getpeername", /* 52 */
	"socketpair", /* 53 */
	"setsockopt", /* 54 */
	"getsockopt", /* 55 */
	"clone", /* 56 */
	"fork", /* 57 */
	"vfork", /* 58 */
	"execve", /* 59 */
	"exit", /* 60 */
	"wait4", /* 61 */
	"kill", /* 62 */
	"uname", /* 63 */
	"semget", /* 64 */
	"semop", /* 65 */
	"semctl", /* 66 */
	"shmdt", /* 67 */
	"msgget", /* 68 */
	"msgsnd", /* 69 */
	"msgrcv", /* 70 */
	"msgctl", /* 71 */
	"fcntl", /* 72 */
	"flock", /* 73 */
	"fsync", /* 74 */
	"fdatasync", /* 75 */
	"truncate", /* 76 */
	"ftruncate", /* 77 */
	"getdents", /* 78 */
	"getcwd", /* 79 */
	"chdir", /* 80 */
	"fchdir", /* 81 */
	"rename", /* 82 */
	"mkdir", /* 83 */
	"rmdir", /* 84 */
	"creat", /* 85 */
	"link", /* 86 */
	"unlink", /* 87 */
	"symlink", /* 88 */
	"readlink", /* 89 */
	"chmod", /* 90 */
	"fchmod", /* 91 */
	"chown", /* 92 */
	"fchown", /* 93 */
	"lchown", /* 94 */
	"umask", /* 95 */
	"gettimeofday", /* 96 */
	"getrlimit", /* 97 */
	"getrusage", /* 98 */
	"sysinfo", /* 99 */
	"times", /* 100 */
	"ptrace", /* 101 */
	"getuid", /* 102 */
	"syslog", /* 103 */
	"getgid", /* 104 */
	"setuid", /* 105 */
	"setgid", /* 106 */
	"geteuid", /* 107 */
	"getegid", /* 108 */
	"setpgid", /* 109 */
	"getppid", /* 110 */
	"getpgrp", /* 111 */
	"setsid", /* 112 */
	"setreuid", /* 113 */
	"setregid", /* 114 */
	"getgroups", /* 115 */
	"setgroups", /* 116 */
	"setresuid", /* 117 */
	"getresuid", /* 118 */
	"setresgid", /* 119 */
	"getresgid", /* 120 */
	"getpgid", /* 121 */
	"setfsuid", /* 122 */
	"setfsgid", /* 123 */
	"getsid", /* 124 */
	"capget", /* 125 */
	"capset", /* 126 */
	"rt_sigpending", /* 127 */
	"rt_sigtimedwait", /* 128 */
	"rt_sigqueueinfo", /* 129 */
	"rt_sigsuspend", /* 130 */
	"sigaltstack", /* 131 */
	"utime", /* 132 */
	"mknod", /* 133 */
	"uselib", /* 134 */
	"personality", /* 135 */
	"ustat", /* 136 */
	"statfs", /* 137 */
	"fstatfs", /* 138 */
	"sysfs", /* 139 */
	"getpriority", /* 140 */
	"setpriority", /* 141 */
	"sched_setparam", /* 142 */
	"sched_getparam", /* 143 */
	"sched_setscheduler", /* 144 */
	"sched_getscheduler", /* 145 */
	"sched_get_priority_max", /* 146 */
	"sched_get_priority_min", /* 147 */
	"sched_rr_get_interval", /* 148 */
	"mlock", /* 149 */
	"munlock", /* 150 */
	"mlockall", /* 151 */
	"munlockall", /* 152 */
	"vhangup", /* 153 */
	"modify_ldt", /* 154 */
	"pivot_root", /* 155 */
	"_sysctl", /* 156 */
	"prctl", /* 157 */
	"arch_prctl", /* 158 */
	"adjtimex", /* 159 */
	"setrlimit", /* 160 */
	"chroot", /* 161 */
	"sync", /* 162 */
	"acct", /* 163 */
	"settimeofday", /* 164 */
	"mount", /* 165 */
	"umount2", /* 166 */
	"swapon", /* 167 */
	"swapoff", /* 168 */
	"reboot", /* 169 */
	"sethostname", /* 170 */
	"setdomainname", /* 171 */
	"iopl", /* 172 */
	"ioperm", /* 173 */
	"create_module", /* 174 */
	"init_module", /* 175 */
	"delete_module", /* 176 */
	"get_kernel_syms", /* 177 */
	"query_module", /* 178 */
	"quotactl", /* 179 */
	"nfsservctl", /* 180 */
	"getpmsg", /* 181 */
	"putpmsg", /* 182 */
	"afs_syscall", /* 183 */
	"tuxcall", /* 184 */
	"security", /* 185 */
	"gettid", /* 186 */
	"readahead", /* 187 */
	"setxattr", /* 188 */
	"lsetxattr", /* 189 */
	"fsetxattr", /* 190 */
	"getxattr", /* 191 */
	"lgetxattr", /* 192 */
	"fgetxattr", /* 193 */
	"listxattr", /* 194 */
	"llistxattr", /* 195 */
	"flistxattr", /* 196 */
	"removexattr", /* 197 */
	"lremovexattr", /* 198 */
	"fremovexattr", /* 199 */
	"tkill", /* 200 */
	"time", /* 201 */
	"futex", /* 202 */
	"sched_setaffinity", /* 203 */
	"sched_getaffinity", /* 204 */
	"set_thread_area", /* 205 */
	"io_setup", /* 206 */
	"io_destroy", /* 207 */
	"io_getevents", /* 208 */
	"io_submit", /* 209 */
	"io_cancel", /* 210 */
	"get_thread_area", /* 211 */
	"lookup_dcookie", /* 212 */
	"epoll_create", /* 213 */
	"epoll_ctl_old", /* 214 */
	"epoll_wait_old", /* 215 */
	"remap_file_pages", /* 216 */
	"getdents64", /* 217 */
	"set_tid_address", /* 218 */
	"restart_syscall", /* 219 */
	"semtimedop", /* 220 */
	"fadvise64", /* 221 */
	"timer_create", /* 222 */
	"timer_settime", /* 223 */
	"timer_gettime", /* 224 */
	"timer_getoverrun", /* 225 */
	"timer_delete", /* 226 */
	"clock_settime", /* 227 */
	"clock_gettime", /* 228 */
	"clock_getres", /* 229 */
	"clock_nanosleep", /* 230 */
	"exit_group", /* 231 */
	"epoll_wait", /* 232 */
	"epoll_ctl", /* 233 */
	"tgkill", /* 234 */
	"utimes", /* 235 */
	"vserver", /* 236 */
	"mbind", /* 237 */
	"set_mempolicy", /* 238 */
	"get_mempolicy", /* 239 */
	"mq_open", /* 240 */
	"mq_unlink", /* 241 */
	"mq_timedsend", /* 242 */
	"mq_timedreceive", /* 243 */
	"mq_notify", /* 244 */
	"mq_getsetattr", /* 245 */
	"kexec_load", /* 246 */
	"waitid", /* 247 */
	"add_key", /* 248 */
	"request_key", /* 249 */
	"keyctl", /* 250 */
	"ioprio_set", /* 251 */
	"ioprio_get", /* 252 */
	"inotify_init", /* 253 */
	"inotify_add_watch", /* 254 */
	"inotify_rm_watch", /* 255 */
	"migrate_pages", /* 256 */
	"openat", /* 257 */
	"mkdirat", /* 258 */
	"mknodat", /* 259 */
	"fchownat", /* 260 */
	"futimesat", /* 261 */
	"newfstatat", /* 262 */
	"unlinkat", /* 263 */
	"renameat", /* 264 */
	"linkat", /* 265 */
	"symlinkat", /* 266 */
	"readlinkat", /* 267 */
	"fchmodat", /* 268 */
	"faccessat", /* 269 */
	"pselect6", /* 270 */
	"ppoll", /* 271 */
	"unshare", /* 272 */
	"set_robust_list", /* 273 */
	"get_robust_list", /* 274 */
	"splice", /* 275 */
	"tee", /* 276 */
	"sync_file_range", /* 277 */
	"vmsplice", /* 278 */
	"move_pages", /* 279 */
	"utimensat", /* 280 */
	"ORE_getcpu", /* 0 */
	"epoll_pwait", /* 281 */
	"signalfd", /* 282 */
	"timerfd_create", /* 283 */
	"eventfd", /* 284 */
	"fallocate", /* 285 */
	"timerfd_settime", /* 286 */
	"timerfd_gettime", /* 287 */
	"paccept", /* 288 */
	"signalfd4", /* 289 */
	"eventfd2", /* 290 */
	"epoll_create1", /* 291 */
	"dup3", /* 292 */
	"pipe2", /* 293 */
	"inotify_init1", /* 294 */
	NULL 
};

#endif  /* PTRACE_LINUX64 */

#if !defined(DFPRINTF)
#define DFPRINTF(x)
#endif

enum LINUX_CALL_TYPES {
	LINUX64 = 0,
	LINUX32 = 1,
	LINUX_NUM_VERSIONS = 2
};

static enum LINUX_CALL_TYPES
linux_call_type(long codesegment) 
{
	if (codesegment == 0x33)
		return (LINUX64);
	else if (codesegment == 0x23)
		return (LINUX32);
        else {
		warnx("%s:%d: unknown code segment %lx\n",
		    __FILE__, __LINE__, codesegment);
		assert(0);
	}
}


static int
get_number_syscalls(const char **syscalls)
{
	int i;
	for (i = 0; i < NR_syscalls; ++i) {
		if (syscalls[i] == NULL)
			break;
	}

	return (i);
}

/* returns the system call name based on the system call number */
static const char *
generic_syscall_name(const char **syscallnames, int *pnr_syscalls,
    pid_t pidnr, int number)
{
	int nr_syscalls = *pnr_syscalls;
	/*
	 * Compute the number of available system calls in our translation
	 * table.  This might differ from kernel to kernel.
	 */
	if (nr_syscalls == -1)
		*pnr_syscalls = nr_syscalls = get_number_syscalls(syscallnames);
	if (number < -1 || number >= nr_syscalls * 2) {
		errx(1, "pid %d Bad syscall number: %d\n", pidnr, number);
	}
	/* handle spurious -1 on entry */
	if (number == -1) {
		return (NULL);
	}
	if (number >= nr_syscalls) {
	        /* linux usually has header files and kernel out of sync */
	        static char name[32];
		snprintf(name, sizeof(name), "unknown-%d", number);
		return (name);
	}

	return (syscallnames[number]);
}

static const char *
linux_syscall_name(enum LINUX_CALL_TYPES call_type, pid_t pidnr, int number)
{
#ifdef PTRACE_LINUX64
	static int nr_syscalls_32 = -1;
	static int nr_syscalls_64 = -1;

	switch (call_type) {
	case LINUX32:
		return (generic_syscall_name(
				&linux_syscallnames[0], &nr_syscalls_32,
				pidnr, number));
	case LINUX64:
		return (generic_syscall_name(
				&linux_syscallnames_64[0], &nr_syscalls_64,
				pidnr, number));
	default:
		errx(1, "unknown call_type: %d", call_type);
	};
#else
	static int nr_syscalls = -1;

	return (generic_syscall_name(
			&linux_syscallnames[0], &nr_syscalls, pidnr, number));
#endif
}

static int
linux_syscall_number(const char *emulation, const char *name)
{
	int i;
	const char **syscallnames;

	if (strcmp(emulation, "linux") == 0)
		syscallnames = &linux_syscallnames[0];
#ifdef PTRACE_LINUX64
	else if (strcmp(emulation, "linux64") == 0)
		syscallnames = &linux_syscallnames_64[0];
#endif /* PTRACE_LINUX64 */
	else
		errx(1, "unkown linux emulation: %s", emulation);

	for (i = 0; i < NR_syscalls && syscallnames[i]; i++)
		if (!strcmp(name, syscallnames[i]))
			return i;
	if (strncmp(name, "unknown-", 8) == 0) {
		/* guess the system call number from the generated name */
		return atoi(name + 8);
	}

	return (-1);
}
