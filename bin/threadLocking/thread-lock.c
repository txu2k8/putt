#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>
#include <errno.h>
#include <fcntl.h>
#include <semaphore.h>
#include <sys/mman.h>
#include<pthread.h>
#define _MULTI_THREADED
   #include <pthread.h>
#define NUMTHREADS  100000
#define MAXLOCKS  1000000

 
/* global variables */
/* creating a mutex variable so that multiple threads donot update flock struct */
pthread_mutex_t COUNT_MUTEX = PTHREAD_MUTEX_INITIALIZER;
/*semaphore variable for synchronization */
sem_t SEM0;
char *FILENAME;
char *OPEN_TYPE;
char *LOCK_TYPE;
char *WAIT_TYPE;
long int NUMLOCK;
long int THREADNUM;

/* thread parameters struct */
struct thread_parms
{
  int fd[NUMTHREADS];
  struct flock fl[MAXLOCKS];
};

/* thread function */
void *lockThread(void *parm)
{
  struct thread_parms *tp=parm;;
  int rc = pthread_mutex_lock(&COUNT_MUTEX);
  int start =0;
  static int count=0;
  static int len =10;
  static int i=0;
  int j=0;
  char file[10000];
  short int opentype;
  short int waittype;
  short int locktype;
  printf("-----------------------------------------\n");
  sprintf(file,"%s.%d",FILENAME,i);
  printf("Thread id %u\n",(unsigned int)pthread_self());

  if (atoi(OPEN_TYPE))
  {
    if (atoi(OPEN_TYPE)> 1) 
    {
      opentype=O_WRONLY;
      printf("wronly\n");
    } else {
      opentype=O_RDWR;
      printf("rdwr\n");
    }
  }else {
    opentype=O_RDONLY;
    printf("rdonly\n");
  }

  if (atoi(OPEN_TYPE)== 2) /* for write only*/
  {
     if ((tp->fd[i] = open(file, O_WRONLY | O_CREAT,00200)) == -1) {
           perror("open error\n");
           printf("open error %s\n",file);
           exit(1);
     }
  } else if (atoi(OPEN_TYPE) == 1) { /* for read-write only*/
      if ((tp->fd[i] = open(file, O_RDWR | O_CREAT,00600)) == -1) {
           perror("open error\n");
           printf("open error %s\n",file);
           exit(1);
      }
  } else { /* for read only*/
      if ((tp->fd[i] = open(file, O_RDONLY,00400 )) == -1) {
           perror("open error\n");
           printf("open error %s\n",file);
           exit(1);
      }
  }
 
  printf("Successfully opened the file %s in mode %s\n",file, (opentype > 0) ? ((opentype > 1 ) ? "O_RDWR" : "O_WRONLY") : "O_RDONLY" );

  /* if opened with O_ RDWR mode or O_WRONLY mode*/
  if (opentype)
  {
    printf("Write some data\n");
    if (write(tp->fd[i],  "This will be output\n", 20) == -1) {
      perror("write error:\n");
      exit(1);
    }
  }

  if(atoi(WAIT_TYPE))
    waittype=F_SETLKW; /*blocking*/
  else
    waittype=F_SETLK; /*non-blocking*/

  for (j=0;j<NUMLOCK;j++)
  {
     /* setting file lock properties */
     tp->fl[j].l_type = atoi(LOCK_TYPE);
     tp->fl[j].l_pid = getpid();
     if (atoi(LOCK_TYPE))
     {
       tp->fl[j].l_type = F_WRLCK;
       locktype = F_WRLCK;
     } else {
       tp->fl[j].l_type = F_RDLCK;
       locktype = F_RDLCK;
     }
     tp->fl[j].l_start = start;
     tp->fl[j].l_len = len;
     tp->fl[j].l_whence = SEEK_END;
     printf("Trying to get lock..start %ld len %ld\n",(tp->fl[j].l_start),(tp->fl[j].l_len));
     if (fcntl(tp->fd[i], waittype, &tp->fl[j]) == -1) {
       perror("fcntl error:\n");
       exit(1);
     }
     printf("Lock granted: %s, %s\n", (locktype > 0) ? "F_WRLCK" : "F_RDLCK", (atoi(WAIT_TYPE) > 0) ? "BLOCKING" : "NON-BLOCKING"  );
     start=start+ rand() % 1000; 
  }
  count=count+1;

  printf("Waiting Completed for thread num %d\n",i);
  i=i+1;
  printf("-----------------------------------------\n");
  pthread_mutex_unlock(&COUNT_MUTEX);

  /* if all the threads have acquired byte locks, do a sem post*/
  if (count == THREADNUM) {
    sem_post(&SEM0);
  }

  return NULL;
}

int main(int argc, char **argv)
{
  pthread_t  thread[NUMTHREADS];
  struct thread_parms tp ; 
  int       rc=0;
  long int  i;
  long int  j;

  if (argc != 7) {
          fprintf(stderr, "Usage: %s <FILENAME_postfix> <number of threads> <open type: Example 0 for O_RDONLY or 1 for O_RDWR or 2 for O_WRONLY> <lock-type: 0 for F_RDLCK or  1 for F_WRLCK> <waiting: 0 for F_SETLK or 1 for F_SETLKW> <no of locks per thread>\n",
                      argv[0]);
          return;
  }
 
  FILENAME=argv[1];
  THREADNUM=atoi(argv[2]);  
  OPEN_TYPE=argv[3];
  LOCK_TYPE=argv[4];
  WAIT_TYPE=argv[5];
  NUMLOCK=atoi(argv[6]);

  /* error checking */
  if (THREADNUM<1 || NUMLOCK<1) 
  {
          fprintf(stderr,"THREADNUM and NUMLOCK should be greater than equal to 1\n");
          fprintf(stderr, "Usage: %s <FILENAME_postfix> <number of threads> <open type: Example 0 for O_RDONLY or 1 for O_RDWR or 2 for O_WRONLY> <lock-type: 0 for F_RDLCK or  1 for F_WRLCK> <waiting: 0 for F_SETLK or 1 for F_SETLKW> <no of locks per thread>\n",
                      argv[0]);
          return;
  }
  if (atoi(OPEN_TYPE) >2 || atoi(OPEN_TYPE)<0)
  {
          fprintf(stderr,"OPEN_TYPE should be 0-2\n");
          fprintf(stderr, "Usage: %s <FILENAME_postfix> <number of threads> <open type: Example 0 for O_RDONLY or 1 for O_RDWR or 2 for O_WRONLY> <lock-type: 0 for F_RDLCK or  1 for F_WRLCK> <waiting: 0 for F_SETLK or 1 for F_SETLKW> <no of locks per thread>\n",
                      argv[0]);
          return;
  }
  if (atoi(LOCK_TYPE) >1 || atoi(LOCK_TYPE)<0)
  {
          fprintf(stderr,"LOCK_TYPE should be 0 or 1\n");
          fprintf(stderr, "Usage: %s <FILENAME_postfix> <number of threads> <open type: Example 0 for O_RDONLY or 1 for O_RDWR or 2 for O_WRONLY> <lock-type: 0 for F_RDLCK or  1 for F_WRLCK> <waiting: 0 for F_SETLK or 1 for F_SETLKW> <no of locks per thread>\n",
                      argv[0]);
          return;
  }
  if (atoi(WAIT_TYPE) >1 || atoi(WAIT_TYPE)<0)
  {
          fprintf(stderr,"WAIT_TYPE should be 0 or 1\n");
          fprintf(stderr, "Usage: %s <FILENAME_postfix> <number of threads> <open type: Example 0 for O_RDONLY or 1 for O_RDWR or 2 for O_WRONLY> <lock-type: 0 for F_RDLCK or  1 for F_WRLCK> <waiting: 0 for F_SETLK or 1 for F_SETLKW> <no of locks per thread>\n",
                      argv[0]);
          return;
  }
  if (atoi(OPEN_TYPE) == 0 && atoi(LOCK_TYPE) == 1)  
  {
          fprintf(stderr,"If OPEN_TYPE is Read-Only LOCK_TYPE cannot be Write\n");
          fprintf(stderr, "Usage: %s <FILENAME_postfix> <number of threads> <open type: Example 0 for O_RDONLY or 1 for O_RDWR or 2 for O_WRONLY> <lock-type: 0 for F_RDLCK or  1 for F_WRLCK> <waiting: 0 for F_SETLK or 1 for F_SETLKW> <no of locks per thread>\n",
                      argv[0]);
          return;
  }
  if (atoi(OPEN_TYPE) == 2 && atoi(LOCK_TYPE) == 0)  
  {
          fprintf(stderr,"If OPEN_TYPE is Write-Only LOCK_TYPE cannot be Read\n");
          fprintf(stderr, "Usage: %s <FILENAME_postfix> <number of threads> <open type: Example 0 for O_RDONLY or 1 for O_RDWR or 2 for O_WRONLY> <lock-type: 0 for F_RDLCK or  1 for F_WRLCK> <waiting: 0 for F_SETLK or 1 for F_SETLKW> <no of locks per thread>\n",
                      argv[0]);
          return;
  }

  printf("Initializing an unnamed semaphore\n");
  if ((rc  = sem_init(&SEM0,0,0) ) < 0) {
      perror("sem_destroy failure");
      exit(0);
  }
  printf("Create/start %ld threads\n",THREADNUM);
  for (i=0; i<THREADNUM; i++) {
       if ( (rc = pthread_create(&thread[i], NULL, lockThread, (void *)&tp)) != 0) {
            perror("pthread_create error");
            exit(1);
       }
  }

  /* wait for the semaphore to return */
  while (sem_wait(&SEM0) < 0);
  sleep(10); /*giving some extra time */
  /* Wait for the user to hit Enter. */
  printf ("locked; hit Enter to unlock...\n");
  getchar ();

  printf("Cleanup,Unlock and Close\n");
  for (i=0; i <THREADNUM; i++) {
    for (j=0; j<NUMLOCK; j++) {
      printf("Unlocking for threadnum %ld byte range %ld %ld\n",i,tp.fl[j].l_start,tp.fl[j].l_len); 
      tp.fl[j].l_type = F_UNLCK;  /* set to unlock same region */
      if (fcntl(tp.fd[i], F_SETLK, &tp.fl[j]) == -1) {
         perror("fcntl error\n");
         exit(1);
      }
    }
  }
  for (i=0; i <THREADNUM; i++) {
      if ( close(tp.fd[i]) == -1) {
         perror("close error\n");
      } 
  }

  printf("Destroying the semaphore\n");
  if  ( (rc  = sem_destroy(&SEM0)) < 0) {
      perror("sem_destroy failure");
      exit(1);
  }
 
  printf("Main completed\n");
  return 0;
}
