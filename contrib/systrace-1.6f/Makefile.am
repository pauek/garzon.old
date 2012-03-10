AUTOMAKE_OPTIONS = foreign no-dependencies

SUBDIRS = . regress

DISTCLEANFILES = *~

YFLAGS = -d

bin_PROGRAMS = systrace

systrace_SOURCES =  filter.c intercept-translate.c intercept.c \
		util.c policy.c systrace-errno.h systrace-error.c \
		systrace-translate.c systrace.c alias.c parse.y lex.l \
		cradle.c
systrace_LDADD = @LIBOBJS@ @EVENTLIB@

include_HEADERS =

CFLAGS = -Wall $(AM_CFLAGS)

INCLUDES = @SYSTR_INC@ @EVENTINC@

man_MANS = systrace.1
EXTRA_DIST = $(man_MANS) README openbsd-syscalls.c register-openbsd.c \
	linux-syscalls.c register-linux.c \
	darwin-syscalls.c register-darwin.c \
	sys/queue.h sys/tree.h err.c strlcpy.c strlcat.c \
	lex.c parse.c err.h filter.h intercept.h linux_fcntl.h linux_types.h \
	parse.h systrace-errno.h systrace.h util.h linux_syscalls.c \
	linux-translate.c linux-translate.h \
	linux_socket.h linux_socketcall.h getopt.c setproctitle.c \
	regress/arch.policy.OpenBSD regress/nm.policy.Linux \
	regress/uniq.policy.Linux regress/nm.policy.Linux64 \
	regress/uniq.policy.Linux64 regress/id.policy.NetBSD \
	regress/nm.policy.OpenBSD regress/uniq.policy.OpenBSD \
	regress/id.policy.OpenBSD regress/regress.sh  regress/nm.args \
	regress/uniq.args

