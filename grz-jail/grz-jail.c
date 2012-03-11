
#include <errno.h>
#include <signal.h>
#include <string.h>
#include <sys/time.h>
#include <sys/resource.h>
#include <sys/prctl.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>

void usage() {
   fprintf(stderr, "usage: grz-jail [options...] <exe>\n\n");
   fprintf(stderr, "Options:\n");
   fprintf(stderr, "   -m <mem>   Max megabytes\n");
   fprintf(stderr, "   -t <mem>   Max megabytes\n");
   fprintf(stderr, "   -f <mem>   Max megabytes\n");
   fprintf(stderr, "   -a         Accused mode\n\n");
}

void usr1(int x) {}

void setlimit(int what, rlim_t max) {
   struct rlimit L = { 
      .rlim_cur = max, 
      .rlim_max = max 
   };
   if (max > 0) {
      setrlimit(what, &L);
   }
}

int main(int argc, char *argv[], char *envp[]) {
   int opt;
   int accused = 0;
   int max_cpu_seconds = 2;
   int max_memory = 64 * 1024 * 1024;
   int max_file_size = 1; // 1 byte (?)

   // Parse options
   while ((opt = getopt(argc, argv, "m:t:f:a")) != -1) {
      switch (opt) {
      case 't':
         max_cpu_seconds = atoi(optarg); 
         break;
      case 'm':
         max_memory = atoi(optarg) * 1024 * 1024; 
         break;
      case 'f':
         max_file_size = atoi(optarg) * 1024 * 1024;
         break;
      case 'a':
         accused = 1; break;
      default:
         usage();
         exit(EXIT_FAILURE);
      }
   }

   if ((argc - optind) != 1) {
      fprintf(stderr, "Wrong number of arguments\n");
      usage();
      return EXIT_FAILURE;
   }

   char *exe = argv[optind];
   pid_t pid = getpid();

   pid_t accused_pid = fork();
   if (accused_pid == 0) { // Child
      printf("Creating child process\n");
      setlimit(RLIMIT_CPU,   max_cpu_seconds);
      setlimit(RLIMIT_AS,    max_memory);
      setlimit(RLIMIT_FSIZE, max_file_size);
      signal(SIGUSR1, usr1);
      prctl(PR_SET_PTRACER, pid, 0, 0, 0);
      printf("Child: Sleeping\n");
      sleep(10);
      printf("Child: Resuming\n");
      char *newargv[] = { exe, NULL };
      char *newenv[]  = { NULL };
      if (-1 == execve(exe, newargv, newenv)) {
         fprintf(stderr, "Error executing '%s': %s\n", exe, strerror(errno));
      }
   } 
   printf("Child PID: %d\n", accused_pid);

   pid_t systrace_pid = fork();
   if (systrace_pid == 0) { // systrace
      printf("Creating systrace\n");
      char A[] = "-?", apid[12];
      A[1] = (accused ? 'a' : 'A');
      sprintf(apid, "%d", accused_pid);
      char *newargv[] = { "systrace", A, "-p", apid, exe, NULL };
      if (-1 == execvp("systrace", newargv)) {
         fprintf(stderr, "Error executing 'systrace': %s\n", strerror(errno));
      }
   }
   printf("Systrace PID: %d\n", systrace_pid);
   sleep(1); // Give time to systrace
   kill(accused_pid, SIGUSR1);
   
   int status;
   if (-1 == waitpid(accused_pid, &status, 0)) {
      fprintf(stderr, "Error waiting for child\n");
      return 1;
   }

   if (WIFEXITED(status)) {
      printf("[Exited with status %d]\n", WEXITSTATUS(status));
   } else if (WIFSIGNALED(status)) {
      printf("[Terminated by signal %d]\n", WTERMSIG(status));
   } else {
      printf("[Child exited for unknown reason!]\n");
   }
   
   waitpid(systrace_pid, &status, 0);
}
