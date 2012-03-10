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
	X(intercept_register_sccb("darwin", "open", trans_cb, NULL));
	tl = intercept_register_transfn("darwin", "open", 0);
	intercept_register_translation("darwin", "open", 1, &ic_oflags);
	alias = systrace_new_alias("darwin", "open", "darwin", "fswrite");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_gencb(gen_cb, NULL));
	X(intercept_register_sccb("darwin", "load_shared_file",
	      trans_cb, NULL));
	tl = intercept_register_transfn("darwin", "load_shared_file", 0);
	alias = systrace_new_alias("darwin", "load_shared_file",
	    "darwin", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("darwin", "connect", trans_cb, NULL));
	intercept_register_translation("darwin", "connect", 1,
	    &ic_translate_connect);
	X(intercept_register_sccb("darwin", "sendto", trans_cb, NULL));
	intercept_register_translation("darwin", "sendto", 4,
	    &ic_translate_connect);
	X(intercept_register_sccb("darwin", "bind", trans_cb, NULL));
	intercept_register_translation("darwin", "bind", 1,
	    &ic_translate_connect);
	X(intercept_register_sccb("darwin", "execve", trans_cb, NULL));
	intercept_register_transfn("darwin", "execve", 0);
	intercept_register_translation("darwin", "execve", 1, &ic_trargv);
	X(intercept_register_sccb("darwin", "stat", trans_cb, NULL));
	tl = intercept_register_transfn("darwin", "stat", 0);
	alias = systrace_new_alias("darwin", "stat", "darwin", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("darwin", "lstat", trans_cb, NULL));
	tl = intercept_register_translation("darwin", "lstat", 0,
	    &ic_translate_unlinkname);
	alias = systrace_new_alias("darwin", "lstat", "darwin", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("darwin", "unlink", trans_cb, NULL));
	tl = intercept_register_translation("darwin", "unlink", 0,
	    &ic_translate_unlinkname);
	alias = systrace_new_alias("darwin", "unlink", "darwin", "fswrite");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("darwin", "chown", trans_cb, NULL));
	intercept_register_transfn("darwin", "chown", 0);
	intercept_register_translation("darwin", "chown", 1, &ic_uidt);
	intercept_register_translation("darwin", "chown", 2, &ic_gidt);
	X(intercept_register_sccb("darwin", "fchown", trans_cb, NULL));
	intercept_register_translation("darwin", "fchown", 0, &ic_fdt);
	intercept_register_translation("darwin", "fchown", 1, &ic_uidt);
	intercept_register_translation("darwin", "fchown", 2, &ic_gidt);
	X(intercept_register_sccb("darwin", "chmod", trans_cb, NULL));
	intercept_register_transfn("darwin", "chmod", 0);
	intercept_register_translation("darwin", "chmod", 1, &ic_modeflags);
	X(intercept_register_sccb("darwin", "fchmod", trans_cb, NULL));
	intercept_register_translation("darwin", "fchmod", 0, &ic_fdt);
	intercept_register_translation("darwin", "fchmod", 1, &ic_modeflags);
	X(intercept_register_sccb("darwin", "readlink", trans_cb, NULL));
	tl = intercept_register_translation("darwin", "readlink", 0,
	    &ic_translate_unlinkname);
	alias = systrace_new_alias("darwin", "readlink", "darwin", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("darwin", "chdir", trans_cb, NULL));
	intercept_register_transfn("darwin", "chdir", 0);
	X(intercept_register_sccb("darwin", "chroot", trans_cb, NULL));
	intercept_register_transfn("darwin", "chroot", 0);
	X(intercept_register_sccb("darwin", "access", trans_cb, NULL));
	tl = intercept_register_transfn("darwin", "access", 0);
	alias = systrace_new_alias("darwin", "access", "darwin", "fsread");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("darwin", "mkdir", trans_cb, NULL));
	tl = intercept_register_translation("darwin", "mkdir", 0,
	    &ic_translate_unlinkname);
	alias = systrace_new_alias("darwin", "mkdir", "darwin", "fswrite");
	systrace_alias_add_trans(alias, tl);
	X(intercept_register_sccb("darwin", "rmdir", trans_cb, NULL));
	tl = intercept_register_transfn("darwin", "rmdir", 0);
	alias = systrace_new_alias("darwin", "rmdir", "darwin", "fswrite");
	systrace_alias_add_trans(alias, tl);

	X(intercept_register_sccb("darwin", "rename", trans_cb, NULL));
	intercept_register_translation("darwin", "rename", 0,
	    &ic_translate_unlinkname);
	intercept_register_transfn("darwin", "rename", 1);
	X(intercept_register_sccb("darwin", "symlink", trans_cb, NULL));
	intercept_register_transstring("darwin", "symlink", 0);
	intercept_register_transfn("darwin", "symlink", 1);
	X(intercept_register_sccb("darwin", "link", trans_cb, NULL));
	intercept_register_transfn("darwin", "link", 0);
	intercept_register_transfn("darwin", "link", 1);

	X(intercept_register_sccb("darwin", "setuid", trans_cb, NULL));
	intercept_register_translation("darwin", "setuid", 0, &ic_uidt);
	intercept_register_translation("darwin", "setuid", 0, &ic_uname);
	X(intercept_register_sccb("darwin", "seteuid", trans_cb, NULL));
	intercept_register_translation("darwin", "seteuid", 0, &ic_uidt);
	intercept_register_translation("darwin", "seteuid", 0, &ic_uname);
	X(intercept_register_sccb("darwin", "setgid", trans_cb, NULL));
	intercept_register_translation("darwin", "setgid", 0, &ic_gidt);
	X(intercept_register_sccb("darwin", "setegid", trans_cb, NULL));
	intercept_register_translation("darwin", "setegid", 0, &ic_gidt);

 	X(intercept_register_sccb("darwin", "socket", trans_cb, NULL));
 	intercept_register_translation("darwin", "socket", 0, &ic_sockdom);
 	intercept_register_translation("darwin", "socket", 1, &ic_socktype);
 	X(intercept_register_sccb("darwin", "kill", trans_cb, NULL));
 	intercept_register_translation("darwin", "kill", 0, &ic_pidname);
 	intercept_register_translation("darwin", "kill", 1, &ic_signame);

	X(intercept_register_execcb(execres_cb, NULL));
	X(intercept_register_pfreecb(policyfree_cb, NULL));
}
