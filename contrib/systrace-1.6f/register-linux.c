/*	$OpenBSD: register.c,v 1.10 2002/08/05 14:26:07 provos Exp $	*/
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
#include <sys/tree.h>
#include <limits.h>
#include <stdlib.h>
#include <unistd.h>
#include <stdio.h>
#include <err.h>

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif /* HAVE_CONFIG_H */

#include "linux-translate.h"

#include "intercept.h"
#include "systrace.h"

#define X(x)	if ((x) == -1) \
	err(1, "%s:%d: intercept failed", __func__, __LINE__)

void
systrace_initcb(void)
{
	struct systrace_alias *alias;
	struct intercept_translate *tl;

	X(intercept_init());

	X(intercept_register_gencb(gen_cb, NULL));	
 
	X(intercept_register_sccb("linux", "chown", trans_cb, NULL));
	intercept_register_transfn("linux", "chown", 0);
	intercept_register_translation("linux", "chown", 1, &ic_uidt);
	intercept_register_translation("linux", "chown", 2, &ic_gidt);
	X(intercept_register_sccb("linux", "fchown", trans_cb, NULL));
	intercept_register_translation("linux", "fchown", 0, &ic_fdt);
	intercept_register_translation("linux", "fchown", 1, &ic_uidt);
	intercept_register_translation("linux", "fchown", 2, &ic_gidt);

	X(intercept_register_sccb("linux", "fchmod", trans_cb, NULL));
	intercept_register_translation("linux", "fchmod", 0, &ic_fdt);
	intercept_register_translation("linux", "fchmod", 1, &ic_modeflags);

	X(intercept_register_sccb("linux", "chdir", trans_cb, NULL));
	intercept_register_transfn("linux", "chdir", 0);
	X(intercept_register_sccb("linux", "chroot", trans_cb, NULL));
	intercept_register_transfn("linux", "chroot", 0);

	X(intercept_register_sccb("linux", "setuid", trans_cb, NULL));
	intercept_register_translation("linux", "setuid", 0, &ic_uidt);
	intercept_register_translation("linux", "setuid", 0, &ic_uname);

	X(intercept_register_sccb("linux", "setgid", trans_cb, NULL));
	intercept_register_translation("linux", "setgid", 0, &ic_gidt);

	X(intercept_register_sccb("linux", "open", trans_cb, NULL));
	tl = intercept_register_translink("linux", "open", 0);
	intercept_register_translation("linux", "open", 1, &ic_linux_oflags);
	alias = systrace_new_alias("linux", "open", "linux", "fswrite");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("linux", "stat", trans_cb, NULL));
	tl = intercept_register_translink("linux", "stat", 0);
	alias = systrace_new_alias("linux", "stat", "linux", "fsread");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux", "stat64", trans_cb, NULL));
	tl = intercept_register_translink("linux", "stat64", 0);
	alias = systrace_new_alias("linux", "stat64", "linux", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("linux", "lstat", trans_cb, NULL));
	tl = intercept_register_translink("linux", "lstat", 0);
	alias = systrace_new_alias("linux", "lstat", "linux", "fsread");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux", "lstat64", trans_cb, NULL));
	tl = intercept_register_translink("linux", "lstat64", 0);
	alias = systrace_new_alias("linux", "lstat64", "linux", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("linux", "execve", trans_cb, NULL));
	intercept_register_translink("linux", "execve", 0);
	X(intercept_register_sccb("linux", "access", trans_cb, NULL));
	tl = intercept_register_translink("linux", "access", 0);
	alias = systrace_new_alias("linux", "access", "linux", "fsread");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux", "symlink", trans_cb, NULL));
	intercept_register_transstring("linux", "symlink", 0);
	intercept_register_translink("linux", "symlink", 1);
	X(intercept_register_sccb("linux", "link", trans_cb, NULL));
	intercept_register_translink("linux", "link", 0);
	intercept_register_translink("linux", "link", 1);
	X(intercept_register_sccb("linux", "readlink", trans_cb, NULL));
	tl = intercept_register_translink("linux", "readlink", 0);
	alias = systrace_new_alias("linux", "readlink", "linux", "fsread");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux", "rename", trans_cb, NULL));
	intercept_register_translation("linux", "rename", 0,
	    &ic_translate_unlinkname);
	intercept_register_translink("linux", "rename", 1);
	X(intercept_register_sccb("linux", "mkdir", trans_cb, NULL));
	tl = intercept_register_translation("linux", "mkdir", 0,
	    &ic_translate_unlinkname);
	alias = systrace_new_alias("linux", "mkdir", "linux", "fswrite");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux", "rmdir", trans_cb, NULL));
	tl = intercept_register_translink("linux", "rmdir", 0);
	alias = systrace_new_alias("linux", "rmdir", "linux", "fswrite");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux", "unlink", trans_cb, NULL));
	tl = intercept_register_translink("linux", "unlink", 0);
	alias = systrace_new_alias("linux", "unlink", "linux", "fswrite");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux", "chmod", trans_cb, NULL));
	intercept_register_translink("linux", "chmod", 0);
	intercept_register_translation("linux", "chmod", 1, &ic_modeflags);

 	X(intercept_register_sccb("linux", "fcntl", trans_cb, NULL));
 	intercept_register_translation("linux", "fcntl", 1, &ic_fcntlcmd);

	/* i386 specific translation */
	X(intercept_register_sccb("linux", "old_mmap", trans_cb, NULL));
	intercept_register_translation("linux", "old_mmap", 0,
	    &ic_linux_memprot);

	X(intercept_register_sccb("linux", "mmap2", trans_cb, NULL));
	intercept_register_translation("linux", "mmap2", 2, &ic_memprot);
	X(intercept_register_sccb("linux", "mprotect", trans_cb, NULL));
	intercept_register_translation("linux", "mprotect", 2, &ic_memprot);

	X(intercept_register_sccb("linux", "mknod", trans_cb, NULL));
	intercept_register_translation("linux", "mknod", 0,
	    &ic_translate_unlinkname);
	intercept_register_translation("linux", "mknod", 1, &ic_modeflags);

	X(intercept_register_sccb("linux", "socketcall", trans_cb, NULL));
 	tl = intercept_register_translation("linux", "socketcall", 1, &ic_linux_socket_sockdom);
	alias = systrace_new_alias("linux", "socketcall", "linux", "_socketcall");
	systrace_alias_add_trans(alias, tl);
 	tl = intercept_register_translation("linux", "socketcall", 1, &ic_linux_socket_socktype);
	systrace_alias_add_trans(alias, tl);
 	tl = intercept_register_translation("linux", "socketcall", 1, &ic_linux_connect_sockaddr);
	systrace_alias_add_trans(alias, tl);
 	tl = intercept_register_translation("linux", "socketcall", 1, &ic_linux_sendto_sockaddr);
	systrace_alias_add_trans(alias, tl);
 	tl = intercept_register_translation("linux", "socketcall", 1, &ic_linux_sendmsg_sockaddr);
	systrace_alias_add_trans(alias, tl);
 	tl = intercept_register_translation("linux", "socketcall", 1, &ic_linux_bind_sockaddr);
	systrace_alias_add_trans(alias, tl);
 	tl = intercept_register_translation("linux", "socketcall", 0, &ic_linux_socketcall_catchall);
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("linux", "kill", trans_cb, NULL));
	intercept_register_translation("linux", "kill", 0, &ic_pidname);
	intercept_register_translation("linux", "kill", 1, &ic_signame);

#ifdef PTRACE_LINUX64
	X(intercept_register_sccb("linux64", "chown", trans_cb, NULL));
	intercept_register_transfn("linux64", "chown", 0);
	intercept_register_translation("linux64", "chown", 1, &ic_uidt);
	intercept_register_translation("linux64", "chown", 2, &ic_gidt);
	X(intercept_register_sccb("linux64", "fchown", trans_cb, NULL));
	intercept_register_translation("linux64", "fchown", 0, &ic_fdt);
	intercept_register_translation("linux64", "fchown", 1, &ic_uidt);
	intercept_register_translation("linux64", "fchown", 2, &ic_gidt);

	X(intercept_register_sccb("linux64", "fchmod", trans_cb, NULL));
	intercept_register_translation("linux64", "fchmod", 0, &ic_fdt);
	intercept_register_translation("linux64", "fchmod", 1, &ic_modeflags);

	X(intercept_register_sccb("linux64", "chdir", trans_cb, NULL));
	intercept_register_transfn("linux64", "chdir", 0);
	X(intercept_register_sccb("linux64", "chroot", trans_cb, NULL));
	intercept_register_transfn("linux64", "chroot", 0);

	X(intercept_register_sccb("linux64", "setuid", trans_cb, NULL));
	intercept_register_translation("linux64", "setuid", 0, &ic_uidt);
	intercept_register_translation("linux64", "setuid", 0, &ic_uname);

	X(intercept_register_sccb("linux64", "setgid", trans_cb, NULL));
	intercept_register_translation("linux64", "setgid", 0, &ic_gidt);

	X(intercept_register_sccb("linux64", "open", trans_cb, NULL));
	tl = intercept_register_translink("linux64", "open", 0);
	intercept_register_translation("linux64", "open", 1, &ic_linux_oflags);
	alias = systrace_new_alias("linux64", "open", "linux64", "fswrite");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("linux64", "stat", trans_cb, NULL));
	tl = intercept_register_translink("linux64", "stat", 0);
	alias = systrace_new_alias("linux64", "stat", "linux64", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("linux64", "lstat", trans_cb, NULL));
	tl = intercept_register_translink("linux64", "lstat", 0);
	alias = systrace_new_alias("linux64", "lstat", "linux64", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("linux64", "execve", trans_cb, NULL));
	intercept_register_translink("linux64", "execve", 0);
	X(intercept_register_sccb("linux64", "access", trans_cb, NULL));
	tl = intercept_register_translink("linux64", "access", 0);
	alias = systrace_new_alias("linux64", "access", "linux64", "fsread");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux64", "symlink", trans_cb, NULL));
	intercept_register_transstring("linux64", "symlink", 0);
	intercept_register_translink("linux64", "symlink", 1);
	X(intercept_register_sccb("linux64", "link", trans_cb, NULL));
	intercept_register_translink("linux64", "link", 0);
	intercept_register_translink("linux64", "link", 1);
	X(intercept_register_sccb("linux64", "readlink", trans_cb, NULL));
	tl = intercept_register_translink("linux64", "readlink", 0);
	alias = systrace_new_alias("linux64", "readlink", "linux64", "fsread");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux64", "rename", trans_cb, NULL));
	intercept_register_translation("linux64", "rename", 0,
	    &ic_translate_unlinkname);
	intercept_register_translink("linux64", "rename", 1);
	X(intercept_register_sccb("linux64", "mkdir", trans_cb, NULL));
	tl = intercept_register_translation("linux64", "mkdir", 0,
	    &ic_translate_unlinkname);
	alias = systrace_new_alias("linux64", "mkdir", "linux64", "fswrite");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux64", "rmdir", trans_cb, NULL));
	tl = intercept_register_translink("linux64", "rmdir", 0);
	alias = systrace_new_alias("linux64", "rmdir", "linux64", "fswrite");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux64", "unlink", trans_cb, NULL));
	tl = intercept_register_translink("linux64", "unlink", 0);
	alias = systrace_new_alias("linux64", "unlink", "linux64", "fswrite");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("linux64", "chmod", trans_cb, NULL));
	intercept_register_translink("linux64", "chmod", 0);
	intercept_register_translation("linux64", "chmod", 1, &ic_modeflags);

	X(intercept_register_sccb("linux64", "fcntl", trans_cb, NULL));
	intercept_register_translation("linux64", "fcntl", 1, &ic_fcntlcmd);

	X(intercept_register_sccb("linux64", "mmap", trans_cb, NULL));
	intercept_register_translation("linux64", "mmap", 2, &ic_memprot);
	X(intercept_register_sccb("linux64", "mprotect", trans_cb, NULL));
	intercept_register_translation("linux64", "mprotect", 2, &ic_memprot);

	X(intercept_register_sccb("linux64", "mknod", trans_cb, NULL));
	intercept_register_translation("linux64", "mknod", 0,
	    &ic_translate_unlinkname);
	intercept_register_translation("linux64", "mknod", 1, &ic_modeflags);
	
	X(intercept_register_sccb("linux64", "sendmsg", trans_cb, NULL));
	intercept_register_translation("linux64", "sendmsg", 1,
	    &ic_translate_sendmsg);
	X(intercept_register_sccb("linux64", "connect", trans_cb, NULL));
	intercept_register_translation("linux64", "connect", 1,
	    &ic_translate_connect);
	X(intercept_register_sccb("linux64", "sendto", trans_cb, NULL));
	intercept_register_translation("linux64", "sendto", 4,
	    &ic_translate_connect);
	X(intercept_register_sccb("linux64", "bind", trans_cb, NULL));
	intercept_register_translation("linux64", "bind", 1,
	    &ic_translate_connect);

#endif  /* PTRACE_LINUX64 */

	X(intercept_register_execcb(execres_cb, NULL));
	X(intercept_register_pfreecb(policyfree_cb, NULL));
}
