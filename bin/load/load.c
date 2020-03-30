#include<fcntl.h>
#include<unistd.h>
#include<stdio.h>
#include<pthread.h>
#include<errno.h>
#include<string.h>
#include<time.h>
#include<stdlib.h>
#include<sys/stat.h>

#define BLKSZ			4096
#define FSZMB			4095	/*10 GB*/			
#define NUM_THREADS		10	/*keep it a even number*/
#define FQDNLEN			50	/*Len for FQDN for files*/

const char fprefix[6] = "file-";
char *ddpmnt;

/*fill a pre allocated buffer with random data*/
void
__l_createdata (char **bufp,
	size_t size)
{
	int r;
	if (!bufp || !*bufp || sizeof(*bufp) < size)
		return;
	r = rand();
	sprintf(*bufp, "%d", r);
	return;
}

/*delete a file. files names are supposed to have a common prefix with a 
postfix index that makes the name unique*/
void *
l_delfile (void *p)
{
	char fname[255];
	int i;
	i = (int) p;
	sprintf(fname, "%s/%s%d", ddpmnt, fprefix, i);
	printf("Removing %s\n", fname);
	errno = 0;
	if (remove(fname))
		printf("err deleting %s %d %s", fname, errno, strerror(errno));
	return NULL;
}

/*create a file*/
void *
l_createfile (void *p)
{
	char fname[255];
	char *buf;
	int i, fd, blkn = 0;
	long totalblk = 0;
	i = (int) p;
	buf = (char *)malloc(sizeof(char) * BLKSZ);
	if (!buf) {
		printf("%s", strerror(-ENOMEM));
		return NULL;
	}

//	totalblk = (FSZMB * 1024 * 1024)/BLKSZ;
	totalblk = (FSZMB * 1024)/4;
	sprintf(fname, "%s/%s%d", ddpmnt, fprefix, i);
	printf("Creating %s\n", fname);
	errno = 0;
	if ((fd = open(fname, O_CREAT | O_RDWR))) {
		for (blkn = 0; blkn < totalblk ; blkn++) {
			__l_createdata(&buf, sizeof(int));
			errno = 0;
			if (write(fd, buf, BLKSZ) < 0)
				printf("err writing to %s %d %s\n",
					fname, errno, strerror(errno));
		}
		errno = 0;
		if (close(fd))
			printf("err closing %s %d %s\n", fname, errno,
				strerror(errno));
	} else {
		printf("err opening %s %d %s\n", fname, errno, strerror(errno));
	}
	free(buf);
	return NULL;
}

/*creates 10 threads, with 5 threads deleting some files and other 5 writing
data to a different set of files*/
void
main(int argc,
	char *argv[])
{
	pthread_t tid[NUM_THREADS];
	int i, j, k, t = 0;
	void *status;
	srand(time(NULL));

	if (argc < 2) {
		printf ("usage: ./load <dedup mount directory>");
		return;
	}

	ddpmnt = (char *)malloc(sizeof(char) * FQDNLEN);
	if (!ddpmnt) {
		printf("%d %s\n", errno, strerror(errno));
		return;
	}
	ddpmnt = strncpy(ddpmnt, argv[1], FQDNLEN);

	for (i = 0; i < NUM_THREADS/2; i++) {
		pthread_create(&tid[i], NULL, l_createfile, (void *)i);
	}
	for (i = 0; i < NUM_THREADS/2; i++) {
		pthread_join(tid[i], &status);
	}
	j = 0;
	t = 0;
	while (t < 1) {
		/*FIXME all the 5s here are hard coding :(*/
		for (k = 0, i = 5 * j; i < ((5 * j) + 5); i++) {
			pthread_create(&tid[k++], NULL, l_delfile, (void *)i);
		}
		j = j ? 0 : 1;
		for (i = 5 * j; i < ((5 * j) + 5); i++) {
			pthread_create(&tid[k++], NULL, l_createfile, (void *)i);
		}
		for (i = 0; i < NUM_THREADS; i++) {
			pthread_join(tid[i], &status);
		}
		t++;
	}
}
