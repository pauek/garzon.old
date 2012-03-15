
#include <errno.h>
#include <fcntl.h>
#include <signal.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/prctl.h>
#include <sys/ptrace.h>
#include <sys/resource.h>
#include <sys/time.h>
#include <sys/user.h>
#include <sys/wait.h>

int is_accused = 0;
int max_cpu_seconds = 2;
int max_memory = 64 * 1024 * 1024;
int max_file_size = 1024; // 1 Kbyte (for stderr)

pid_t guardian_pid;

void usage_message(char *msg) {
   if (msg != 0) {
      fprintf(stderr, "%s", msg);
   }
   static const char *_usage = 
      "usage: grz-jail [options...] <exe>\n"
      "\n"
      "Options:\n"
      "   -m <mem>   Max megabytes of memory\n"
      "   -t <mem>   Max seconds\n"
      "   -f <mem>   Max megabytes for files\n"
      "   -a         Accused mode\n"
      "\n";
   fprintf(stderr, "%s", _usage);
   exit(1);
}

#define FORMAT __attribute__((format(printf,2,3)))

void kill_accused();

void FORMAT __die(int code, char *msg, ...) {
   kill_accused();
   va_list args;
   va_start(args, msg);
   vfprintf(stderr, msg, args);
   exit(2);
}

#define report(...) __die(0, __VA_ARGS__)
#define die(...)    __die(2, __VA_ARGS__)
#define die_if(cond, ...) if (cond) __die(2, __VA_ARGS__)

void setlimit(int what, rlim_t max) {
   if (max > 0) {
      struct rlimit L = { .rlim_cur = max, .rlim_max = max + 1 };
      if (setrlimit(what, &L) < 0) {
         die("setrlimit(%d, ...)\n", what);
      }
   }
}

inline int milliseconds(struct timeval *t) {
   return t->tv_sec * 1000 + t->tv_usec / 1000;
}

/** Accused **/

pid_t accused_pid = 0;
int   passed_exec = 0;
struct timeval start_time;
struct rusage usage;
struct user user;

typedef struct _syscall_args {
   uint64_t sys, arg[4], result;
} syscall_args;


void the_accused(int argc, char *argv[]) {
   setlimit(RLIMIT_CPU,   max_cpu_seconds);
   setlimit(RLIMIT_AS,    max_memory);
   setlimit(RLIMIT_FSIZE, max_file_size);
   die_if(ptrace(PTRACE_TRACEME) < 0, "ptrace(PTRACE_TRACEME)\n");
   // redirect stderr (has FSIZE limits!)
   close(2);
   if (open("err", O_WRONLY | O_CREAT | O_TRUNC, 0666) != 2) {
      die("Redirect stderr\n");
   }
   raise(SIGSTOP);
   char *env[] = { NULL };
   execve(argv[0], argv, env);
   die("execve(\"%s\"): %s\n", argv[0], strerror(errno));
}

int wait_for_accused(int *stat) {
   pid_t p = wait4(accused_pid, stat, WUNTRACED, &usage);
   if (p < 0 && errno == EINTR) return 1;
   die_if(p < 0, "wait4 error %d\n", errno);
   die_if(p != accused_pid, "wait4: unknown pid '%d'\n", p);
   return 0;
}

void kill_accused() {
   if (accused_pid > 0) {
      kill(-accused_pid, SIGKILL); // ?
      kill( accused_pid, SIGKILL);
      int p, stat;
      do {
         p = wait4(accused_pid, &stat, 0, &usage);
      } while (p < 0 && errno == EINTR);
      die_if(p < 0, "Lost track of the accused!");
   }
}

inline int final_time() {
   struct timeval total;
   timeradd(&usage.ru_utime, &usage.ru_stime, &total);
   return milliseconds(&total);
}

void accused_exited(int stat) {
   accused_pid = 0;
   if (!passed_exec) {
      report("Internal Error\n");
   }
   int code = WEXITSTATUS(stat);
   if (code != 0) {
      report("Error [%d]\n", code);
   }
   report("Ok [%.3f sec, X MB]\n", final_time() / 1000.0);
}

void accused_signaled(int stat) {
   accused_pid = 0;
   report("Signalled [%d]\n", WTERMSIG(stat));
}

int curr_sys = -1;
int syscall_count = 0;

#define NATIVE_NR_execve 59 /* 64-bit execve */

void get_syscall_args(syscall_args *args, int after) {
   int ret = ptrace(PTRACE_GETREGS, accused_pid, NULL, &user);
   die_if(ret < 0, "ptrace(PTRACE_GETREGS)\n");
   args->sys = user.regs.orig_rax;
   args->result = user.regs.rax;
   if (after) return;
   // OJO: Asumimos (sys_type == 64) en box.c original
   // TODO: Protección de syscalls de 32-bits en modo 64-bits
   args->arg[1] = user.regs.rdi;
   args->arg[2] = user.regs.rsi;
   args->arg[3] = user.regs.rdx;
}

void accused_before_syscall() {
   syscall_args args;
   get_syscall_args(&args, 0);
   curr_sys = args.sys;
   if (!passed_exec) {
      if (args.sys == NATIVE_NR_execve) {
         passed_exec = 1;
         return;
      }
   }
   syscall_count++;

   // - Filtrar las syscalls no permitidas
   // - Almacenar el syscall
}

void accused_after_syscall() {
   syscall_args args;
   get_syscall_args(&args, 1);
   if (args.sys == ~(uint64_t)0) {
      // Check return value? Why?
   } else {
      if (args.sys != curr_sys) {
         report("Mismatched syscall before/after\n");
      }
   }
   curr_sys = -1;
}

void accused_stopped(int stat) {
   static int stop_count = 0, sys_tick = 0;

   int sig = WSTOPSIG(stat);

   if (sig == SIGSTOP) {
      // first signal
      int ret = ptrace(PTRACE_SETOPTIONS, accused_pid, NULL, 
                       (void*) PTRACE_O_TRACESYSGOOD);
      die_if(ret < 0, "ptrace(PTRACE_SETOPTIONS)");
   } else if (sig == (SIGTRAP | 0x80)) {  // Syscall
      if (++sys_tick & 1) { // Syscall entry
         accused_before_syscall();
      } else {
         accused_after_syscall();
      }
   } else {
      switch (sig) {
      case SIGABRT: report("Aborted\n");
      case SIGSEGV: report("Segmentation Fault\n");
      case SIGXCPU: report("Time-Limit Exceeded\n");
      case SIGXFSZ: report("File-Size Exceeded\n");
      case SIGTRAP: 
         if (++stop_count > 1) {
            report("Breakpoint\n");
         }
      }
   }
   ptrace(PTRACE_SYSCALL, accused_pid, 0, 0);
}

inline void get_start_time() {
   gettimeofday(&start_time, NULL);
}

int ellapsed_time_ms() {
   struct timeval now, wall;
   gettimeofday(&now, NULL);
   timersub(&now, &start_time, &wall);
   return wall.tv_sec * 1000 + wall.tv_usec/1000;
}


/** Guardian **/

void guardian() {
   int stat;
   
   // signal(INT)
   // signal(ALRM)

   get_start_time();
   
   while (1) {
      int cont = wait_for_accused(&stat);
      if (cont) {
         continue;
      } else if (WIFEXITED(stat)) {
         accused_exited(stat);
      } else if (WIFSIGNALED(stat)) {
         accused_signaled(stat);
      } else if (WIFSTOPPED(stat)) {
         accused_stopped(stat);
      } else {
         die("wait4: unknown status '%d'", stat);
      }
   }
}

int main(int argc, char *argv[]) {
   int opt;

   while ((opt = getopt(argc, argv, "m:t:f:a")) != -1) {
      switch (opt) {
      case 't': max_cpu_seconds = atoi(optarg); break;
      case 'm': max_memory = atoi(optarg) * 1024 * 1024; break;
      case 'f': max_file_size = atoi(optarg) * 1024 * 1024; break;
      case 'a': is_accused = 1; break;
      default:
         usage_message(0);
         exit(EXIT_FAILURE);
      }
   }

   if ((argc - optind) != 1) {
      usage_message("Wrong number of arguments\n");
   }

   accused_pid = fork();
   die_if(accused_pid < 0, "Couldn't fork\n");
   if (accused_pid == 0) { // Child
      the_accused(argc - optind, argv + optind);
   } else {
      guardian();
   }
   die("You entered The Matrix...\n");
   return 3;
}

/* Local variables: */
/* compile-command: "gcc -Wall -static -o grz-jail grz-jail.c" */
/* End: */
