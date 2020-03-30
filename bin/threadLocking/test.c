#include <unistd.h>
#include <fcntl.h>
#include <stdio.h>
 
int main(int av, char **ac)
{
    struct flock fl;
    if (av < 2) {
                fprintf(stderr, "Usage: %s <filename> \n",
                        ac[0]);
                return;
    }
    char *file;
    int fd ;
    file = ac[1];
    /*if (fd = open(file, O_WRONLY|O_CREAT|O_TRUNC, 0644) == -1)
        printf("error opening \n");
    close(fd);*/
     
    //if ((fd = open(file, O_WRONLY )) == -1) {
    //if ((fd = open(file, O_RDWR )) == -1) {
    if ((fd = open(file, O_RDONLY )) == -1) {
      fprintf(stderr, "open error %s\n", file);
      return ;
    }
    fl.l_start = 0L;
    fl.l_whence = 0L;
    fl.l_len = 0L;
    fl.l_type = F_RDLCK;
    //fl.l_type = F_WRLCK;

    printf("Trying to get lock..start %ld len %ld\n",fl.l_start,fl.l_len);
    if (fcntl(fd,F_SETLK , &fl) == -1) {
      perror("fcntl lock error:\n");
      return 0;
    }
    getchar();

    printf("Lock Granted\n");
    fl.l_type = F_UNLCK;
    if (fcntl(fd, F_SETLK, &fl) == -1) {
      perror("fcntl unlock error:\n");
      return 0;
    }

    close(fd);

    return 0;
}
