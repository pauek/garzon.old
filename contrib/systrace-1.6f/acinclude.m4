dnl from Dug Song <dugsong@monkey.org>
dnl
dnl Check for 4.4 BSD sa_len member in sockaddr struct
dnl
dnl usage:      AC_DNET_SOCKADDR_SA_LEN
dnl results:    HAVE_SOCKADDR_SA_LEN (defined)
dnl
AC_DEFUN(AC_DNET_SOCKADDR_SA_LEN,
    [AC_MSG_CHECKING(for sa_len in sockaddr struct)
    AC_CACHE_VAL(ac_cv_dnet_sockaddr_has_sa_len,
        AC_TRY_COMPILE([
#       include <sys/types.h>
#       include <sys/socket.h>],
        [u_int i = sizeof(((struct sockaddr *)0)->sa_len)],
        ac_cv_dnet_sockaddr_has_sa_len=yes,
        ac_cv_dnet_sockaddr_has_sa_len=no))
    AC_MSG_RESULT($ac_cv_dnet_sockaddr_has_sa_len)
    if test $ac_cv_dnet_sockaddr_has_sa_len = yes ; then
            AC_DEFINE(HAVE_SOCKADDR_SA_LEN, 1,
                      [Define if sockaddr struct has sa_len.])
    fi])


