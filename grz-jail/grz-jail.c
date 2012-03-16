
/* 
   Garzón Jail                               (c) 2012, Pau Fernández
   -----------

   grz-jail is a sandbox to execute programs submitted to the Garzón
   Judge system. It is different from other sandboxes because in
   grz-jail, instead of deciding beforehand what syscalls to filter,
   the execution of a "model" program serves as a filter for the
   "accused" program (the one being judged). That is, the "accused"
   program can only execute the syscalls that the "model" did. 

   In this way, you have more flexibility to specify, for each
   problem, what is permitted and what is not (provided that the
   solutions, i.e. the "models", for each problem are trusted).

   To achieve this, grz-jail has two modes:

   - In "model mode", grz-jail generates a file which contains string
     representations of the syscalls the child process has made during
     execution. 

   - In "accused mode" (-a), grz-jail reads the list of syscalls'
     representations from a file and only allows the child process to
     execute those syscalls whose exact representation is found on the
     list.

   The representation of syscalls as strings is, therefore,
   crucial. Since some syscalls have arguments that are adresses and
   depend explicitly on the location of the process in memory, they
   usually do not appear in the string representation. However,
   filenames do appear, since they are sensitive to be allowed or
   denied specifically. For example, an 'open' system call has a path
   as a first argument, and the string representation is:

       open("/tmp/data")
   
   Since the path is embedded in the string representation, if this
   string is in the list of permitted syscalls, only 'open' calls that
   have exactly the same path will succeed in "accused mode".

   Acknowledgements
   ----------------

   The code below is heavily inspired by the box.c in the MO-Eval
   distribution, by Martin Mares (http://mj.ucw.cz/mo-eval/).

*/


#define _LARGEFILE64_SOURCE
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
#include <sys/signal.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/user.h>
#include <sys/wait.h>

int accused_mode = 0;
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

#define SIGNAL(x) __NR_##x
#define sizeof_array(A) (int)(sizeof(A)/sizeof(A[0]))

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

#define PROC_BUF_SIZE 4096
void read_proc_file(pid_t pid, char *buf, char *name, int *fdp) {
   /* Taken from box.c almost unchanged */
   if (!*fdp) {
      sprintf(buf, "/proc/%d/%s", (int) pid, name);
      *fdp = open(buf, O_RDONLY);
      die_if(*fdp < 0, "open(%s): %s\n", buf, strerror(errno));
   }
   lseek(*fdp, 0, SEEK_SET);
   int c = read(*fdp, buf, PROC_BUF_SIZE-1);
   die_if(c < 0, "read on /proc/$pid/%s: %s\n", name, strerror(errno));
   die_if(c >= PROC_BUF_SIZE-1, "/proc/$pid/%s too long\n", name);
   buf[c] = 0;
}

void read_proc_status(pid_t pid, char *buf) {
   static int proc_status_fd;
   read_proc_file(pid, buf, "status", &proc_status_fd);
}

int read_user_mem(pid_t pid, uint64_t addr, char *buf, int len) {
   /* Taken from box.c almost unchanged */
   static int mem_fd;
   if (!mem_fd) {
      char memname[64];
      sprintf(memname, "/proc/%d/mem", (int)pid);
      mem_fd = open(memname, O_RDONLY);
      if (mem_fd < 0)
         die("open(%s): %m", memname);
   }
   if (lseek(mem_fd, addr, SEEK_SET) < 0) {
      die("lseek(mem): %m");
   }
   return read(mem_fd, buf, len);
}

// syscall tables

#define __SYSCALL(a, b) [a] = #b,
   const char *_syscall_names[] = {
#include <asm/unistd.h>
   };
#undef __SYSCALL

inline const char *syscall_name(unsigned int id) {
   if (id < sizeof_array(_syscall_names)) {
      return _syscall_names[id] + 4; // +4 to remove "sys_"
   } else {
      return NULL;
   }
}

#define NUM_SYSCALLS sizeof_array(_syscall_names)

static const char* _syscall_arg_types[NUM_SYSCALLS + 64] = { // +64?
#define S(x) [__NR_##x]
   // Syscalls with filenames in them
   S(open)     = "f*",
   S(creat)    = "f*",
   S(unlink)   = "f",
   S(access)   = "f*", 
   S(truncate) = "f*",
   S(stat)     = "f*",
   S(lstat)    = "f*",
   S(readlink) = "f*",
   S(chmod)    = "fi",

   // Syscalls with file descriptors
   S(read)      = "i..",
   S(write)     = "i..",
   S(close)     = "i",
   S(lseek)     = "i..",
   S(dup)       = "i",
   S(dup2)      = "ii",
   S(ftruncate) = "i.",
   S(fstat)     = "i.",
   S(readv)     = "i..",
   S(writev)    = "i..",
   S(pread64)   = "i...",
   S(pwrite64)  = "i...",
   S(fcntl)     = "ii*",
   S(ioctl)     = "ii",
   S(fchmod)    = "ii",

   // Others
   S(exit)            = "i",
   S(exit_group)      = "i",
   S(arch_prctl)      = "i.",
   S(getpid)          = "",
   S(getuid)          = "",
   S(brk)             = ".",
   S(personality)     = "i",
   S(getresuid)       = "*",
   S(mmap)            = "*",
   S(munmap)          = "*",
   S(uname)           = ".",
   S(gettid)          = "",
   S(set_thread_area) = ".",
   S(get_thread_area) = ".",
   S(set_tid_address) = ".",
   S(time)            = ".",
   S(alarm)           = "i",
   S(pause)           = "",
   S(nanosleep)       = "*",
#undef S
};

inline const char *syscall_arg_types(unsigned int id) {
   if (id < sizeof_array(_syscall_arg_types)) {
      return _syscall_arg_types[id];
   } else {
      return NULL;
   }
}

/** Accused **/

pid_t accused_pid = 0;
int   passed_exec = 0;
int   accused_mem_peak_kb = 0;
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
   if (2 != open("/dev/null", O_WRONLY | O_APPEND)) {
      die("Redirect stderr to '/dev/null'\n");
   }
   raise(SIGSTOP);
   char *env[] = { NULL };
   execve(argv[0], argv, env);
   die("execve(\"%s\"): %s\n", argv[0], strerror(errno));
}

void accused_sample_mem_peak() {
   /* Taken from box.c almost unchanged */

   /*
    *  We want to find out the peak memory usage of the process, which
    *  is maintained by the kernel, but unforunately it gets lost when
    *  the process exits (it is not reported in struct
    *  rusage). Therefore we have to sample it whenever we suspect
    *  that the process is about to exit.
    */
   char buf[PROC_BUF_SIZE], *x;
   read_proc_status(accused_pid, buf);
   
   x = buf;
   while (*x) {
      char *key = x;
      while (*x && *x != ':' && *x != '\n') x++;
      if (!*x || *x == '\n') break;
      *x++ = 0;
      while (*x == ' ' || *x == '\t') x++;
      char *val = x;
      while (*x && *x != '\n') x++;
      if (!*x) break;
      *x++ = 0;
      if (!strcmp(key, "VmPeak")) {
         int peak = atoi(val);
         if (peak > accused_mem_peak_kb)
            accused_mem_peak_kb = peak;
      }
   }
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
      accused_sample_mem_peak();
      ptrace(PTRACE_KILL, accused_pid);
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
   report("Ok [%.3f sec, %.3f MB]\n", 
          final_time() / 1000.0, 
          accused_mem_peak_kb / 1024.0);
}

void accused_signaled(int stat) {
   accused_pid = 0;
   report("Signalled [%d]\n", WTERMSIG(stat));
}

int curr_sys = -1;

const char *get_syscall_filename_arg(uint64_t addr) {
   static char namebuf[4096];
   char *p = namebuf, *end = namebuf;
   do {
      if (p >= end) {
         int remains = PAGE_SIZE - (addr & (PAGE_SIZE-1));
         int l = namebuf + sizeof(namebuf) - end;
         if (l > remains) l = remains;
         if (!l) report("FA: Access to file with name too long\n");
         remains = read_user_mem(accused_pid, addr, end, l);
         die_if(remains < 0, "read(mem): %s\n", strerror(errno));
         if (!remains) {
            report("FA: Access to file with name out of memory\n");
         }
         end += remains;
         addr += remains;
      }
   } while (*p++);
   return namebuf;
}

void get_syscall_args(syscall_args *args, int after) {
   int ret = ptrace(PTRACE_GETREGS, accused_pid, NULL, &user);
   die_if(ret < 0, "ptrace(PTRACE_GETREGS)\n");
   args->sys = user.regs.orig_rax;
   args->result = user.regs.rax;
   if (after) return;
   // OJO: Asumimos (sys_type == 64) en box.c original
   // TODO: Protección de syscalls de 32-bits en modo 64-bits???
   args->arg[1] = user.regs.rdi;
   args->arg[2] = user.regs.rsi;
   args->arg[3] = user.regs.rdx;
}

/*
  Información sobre syscalls:
  http://www.lxhp.in-berlin.de/lhpsysc0.html
*/

char *syscall_to_string(syscall_args *args) {
   static char repr[4096];
   char *cur = repr;

   const char *name = syscall_name(args->sys);
   if (name == NULL) name = "?";

   intmax_t arg[] = { args->arg[1], args->arg[2], args->arg[3] };

   const char *types = syscall_arg_types(args->sys);
   if (types == NULL) types = "___";
   int i, len = strlen(types);
   
   cur += sprintf(cur, "%s(", name);
   for (i = 0; i < len; i++) {
      if (types[i] == '*') break;
      if (i > 0) *cur++ = ',';
      switch (types[i]) {
      case 'i': cur += sprintf(cur, "%lu", (uint64_t)arg[i]); break;
      case 'f': cur += sprintf(cur, "\"%s\"", get_syscall_filename_arg(arg[i])); break;
      case '.': cur += sprintf(cur, "_"); break;
      default:  cur += sprintf(cur, "%lx", arg[i]);
      };
   }
   cur += sprintf(cur, ")");
   return repr;
}

void accused_before_syscall() {
   syscall_args args;
   get_syscall_args(&args, 0);
   curr_sys = args.sys;
   if (!passed_exec) {
      if (args.sys == SIGNAL(execve)) {
         passed_exec = 1;
         return;
      }
   }

   // Maybe sample mem peak
   if (args.sys == SIGNAL(exit) ||
       args.sys == SIGNAL(exit_group)) {
      accused_sample_mem_peak();
   }
   
   fprintf(stderr, "%s\n", syscall_to_string(&args));

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
      case SIGINT:  report("Interrupted\n");
      case SIGILL:  report("Illegal Instruction\n");
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

void accused_check_exe(char *argv0) {
   struct stat _stat;
   die_if(stat(argv0, &_stat) == -1, "Cannot find executable '%s'\n", argv0);
}


/** Guardian **/

void guardian() {
   int stat;
   
   // signal(INT)

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
   while (-1 != (opt = getopt(argc, argv, "m:t:f:a"))) {
      switch (opt) {
      case 't': max_cpu_seconds = atoi(optarg); break;
      case 'm': max_memory = atoi(optarg) * 1024 * 1024; break;
      case 'f': max_file_size = atoi(optarg) * 1024 * 1024; break;
      case 'a': accused_mode = 1; break;
      default: usage_message(0);
      }
   }
   argv += optind;
   argc -= optind;

   if (argc < 1) {
      usage_message("Wrong number of arguments\n");
   }

   accused_check_exe(argv[0]);

   accused_pid = fork();
   die_if(accused_pid < 0, "Couldn't fork\n");
   if (accused_pid == 0) { // Child
      the_accused(argc, argv);
   } else {
      // fprintf(stderr, "Accused PID = %d\n", accused_pid);
      guardian();
   }
   die("Internal Error\n");
   return 3;
}

/* Local variables: */
/* compile-command: "gcc -Wall -static -o grz-jail grz-jail.c" */
/* End: */
