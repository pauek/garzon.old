dnl $Id: configure.in,v 1.29 2006/04/24 05:19:40 provos Exp $
AC_INIT(systrace.h)

AC_CANONICAL_SYSTEM

AM_INIT_AUTOMAKE(systrace, 1.6f)
AM_CONFIG_HEADER(config.h)

AC_PROG_MAKE_SET

dnl Checks for programs.
AC_PROG_CC
AC_PROG_RANLIB
AC_PROG_INSTALL

AM_PROG_LEX
AC_PROG_YACC

dnl Checks for header files.
AC_HEADER_STDC
AC_HEADER_SYS_WAIT

dnl XXX sys/{queue,tree}.h ?
AC_CHECK_HEADERS(sys/ioctl.h sys/time.h unistd.h poll.h sched.h linux/sched.h linux/systrace.h linux/user.h sys/user.h asm/user.h sys/prctl.h libgen.h sched.h)

AC_DNET_SOCKADDR_SA_LEN

SYSTR_INC=""

dnl Platform specific stuff
case "$host" in
*-*-openbsd*)
	AC_LIBOBJ(openbsd-syscalls)
	AC_LIBOBJ(register-openbsd)
	AC_LIBOBJ(linux-translate)
	SYSTR_HEADER="<dev/systrace.h>"
	SYSTR_INC="-I/sys"
        ;;
*-*-netbsd*)
	AC_LIBOBJ(netbsd-syscalls)
	AC_LIBOBJ(register-openbsd)
	AC_LIBOBJ(linux-translate)
	SYSTR_HEADER="<sys/systrace.h>"
	SYSTR_INC="-I/sys"
        ;;
dnl detecting 64-bit Linux
x86_64-*-linux*)
	AC_LIBOBJ(register-linux)
	AC_LIBOBJ(linux-translate)
	AC_MSG_CHECKING([for secure systrace implementation])
	AC_MSG_RESULT([no, using insecure ptrace])
	AC_LIBOBJ(linux-ptrace-syscalls)
	SYSTR_HEADER="<sys/ptrace.h>"
	AC_DEFINE(PTRACE_LINUX64, 1,
	[Define if we are compiling against ptrace on a 64-bit system])
	;;
dnl this catches all other Linux implementations
*-*-linux*)
	AC_LIBOBJ(register-linux)
	AC_LIBOBJ(linux-translate)
	AC_MSG_CHECKING([for secure systrace implementation])
	if test "x$ac_cv_header_linux_systrace_h" = "xyes"; then
		AC_MSG_RESULT([yes])
		AC_LIBOBJ(linux-syscalls)
		SYSTR_HEADER="<linux/systrace.h>"
	else
		AC_MSG_RESULT([no - using insecure ptrace])
		AC_LIBOBJ(linux-ptrace-syscalls)
		AC_DEFINE(PTRACE_LINUX32, 1,
		[Define if we are compiling against ptrace on a 32-bit system])
		SYSTR_HEADER="<sys/ptrace.h>"
	fi
	;;
*-*-darwin*)
	AC_LIBOBJ(darwin-syscalls)
	AC_LIBOBJ(register-darwin)
	SYSTR_HEADER="<sys/systrace.h>"
	;;
*)
	echo "Your platform is currently not supported"
	exit
	;;
esac
AC_SUBST(SYSTR_INC)

dnl Checks for typedefs, structures, and compiler characteristics.
AC_C_CONST
AC_TYPE_PID_T
AC_TYPE_SIZE_T
AC_HEADER_TIME
AC_CHECK_TYPE(u_int64_t, unsigned long long)
AC_CHECK_TYPE(u_int32_t, unsigned int)
AC_CHECK_TYPE(u_int16_t, unsigned short)
AC_CHECK_TYPE(u_int8_t, unsigned char)
AC_CHECK_TYPE(in_addr_t, unsigned long)
AC_CHECK_TYPE(socklen_t, unsigned int)

dnl Checks for library functions.
AC_PROG_GCC_TRADITIONAL
AC_TYPE_SIGNAL
AC_CHECK_FUNCS(select poll socket chflags prctl)
AC_REPLACE_FUNCS(strlcpy strlcat basename dirname setproctitle getopt)

if test "x$ac_cv_func_basename" = "xyes"; then
	AC_MSG_CHECKING(for working basename)
	AC_TRY_RUN(
#include <sys/types.h>
#include <string.h>
int
main(int argc, char **argv)
{
	char *name = "solaris/sucks";
	char *p;
	int len = strlen(name);
	p = basename(name);
	exit(len != strlen(name) || p == name + strlen("solaris/"));
}, AC_MSG_RESULT(yes), [
	AC_MSG_RESULT(no)
	AC_LIBOBJ(basename)
], AC_MSG_RESULT(yes))
fi

if test "x$ac_cv_func_dirname" = "xyes"; then
	AC_MSG_CHECKING(for working dirname)
	AC_TRY_RUN(
#include <sys/types.h>
#include <string.h>
int
main(int argc, char **argv)
{
	char *name = "solaris/sucks";
	char *p;
	int len = strlen(name);
	p = dirname(name);
	exit(len != strlen(name));
}, AC_MSG_RESULT(yes), [
	AC_MSG_RESULT(no)
	AC_LIBOBJ(dirname)
], AC_MSG_RESULT(yes))
fi

dnl Checks for libevent
AC_MSG_CHECKING(for libevent)
AC_ARG_WITH(libevent,
[  --with-libevent=DIR     use libevent in DIR],
[ case "$withval" in
  yes|no)
     AC_MSG_RESULT(no)
     ;;
  *)
     AC_MSG_RESULT($withval)
     if test -f $withval/include/event.h -a -f $withval/lib/libevent.a; then
        owd=`pwd`
        if cd $withval; then withval=`pwd`; cd $owd; fi
        EVENTINC="-I$withval/include"
        EVENTLIB="-L$withval/lib -levent"
     elif test -f $withval/event.h -a -f $withval/libevent.a; then
        owd=`pwd`
        if cd $withval; then withval=`pwd`; cd $owd; fi
        EVENTINC="-I$withval"
        EVENTLIB="-L$withval -levent"
     else
        AC_ERROR(event.h or libevent.a not found in $withval)
     fi
     ;;
  esac ],
[ if test -f ${prefix}/include/event.h -a -f ${prefix}/lib/libevent.a; then
     EVENTINC="-I${prefix}/include"
     EVENTLIB="-L${prefix}/lib -levent"
  elif test -f /usr/local/include/event.h -a -f /usr/local/lib/libevent.a; then
     EVENTINC="-I/usr/local/include"
     EVENTLIB="-L/usr/local/lib -levent"
  elif test -f /usr/include/event.h; then
     EVENTLIB="-levent"
  else
     AC_MSG_RESULT(no)
     AC_ERROR(libevent not found)
  fi
  AC_MSG_RESULT(yes) ]
)
AC_SUBST(EVENTINC)
AC_SUBST(EVENTLIB)

AC_MSG_CHECKING([for symlink race condition protection])
AC_TRY_COMPILE(
[
#include <sys/param.h>
#include <sys/types.h>
#include $SYSTR_HEADER
],
[ struct systrace_replace s; s.strr_flags[0] = 1; ],
	[ AC_DEFINE(HAVE_STRR_FLAGS, 1,
		[Define if systrace_replace has strr_flags])
	  AC_MSG_RESULT([yes])], AC_MSG_RESULT([no])
)

AC_MSG_CHECKING([for st_mtimespec in struct stat])
AC_TRY_COMPILE(
[
#include <sys/param.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
],
[ struct stat s; s.st_mtimespec.tv_sec = 0; ],
	[ AC_DEFINE(HAVE_ST_MTIMSPEC, 1,	 
		[Define if struct stat has member st_mtimespec])
	  AC_MSG_RESULT([yes])], AC_MSG_RESULT([no])
)

AC_MSG_CHECKING([for execve messages])
AC_TRY_COMPILE(
[
#include <sys/param.h>
#include <sys/types.h>
#include $SYSTR_HEADER
],
[ SYSTR_MSG_EXECVE; ],
	[ AC_DEFINE(HAVE_MSG_EXECVE, 1,
		[Define if systrace has execve messages])
	  AC_MSG_RESULT([yes])], AC_MSG_RESULT([no])
)


neederr=no
AC_CHECK_FUNCS(warnx, , [neederr=yes])
if test $neederr = yes; then
   AC_LIBOBJ(err)
fi

AC_OUTPUT(Makefile regress/Makefile)
