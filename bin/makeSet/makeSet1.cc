// makeSet1.cc - Make regression data set

#ifdef WIN32
#include <Windows.h>
#include <WinError.h>
#include <tchar.h>
#include <io.h>
#include <direct.h>
#endif
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <errno.h>
#ifndef WIN32
#include <sys/time.h>
#include <unistd.h>
#include <sys/types.h>
#include <dirent.h>
#endif
#include <math.h>
#include <sys/stat.h>
#ifdef linux
#include <stdint.h>			// Included for uint32_t type
#endif

#ifdef WIN32
#pragma warning(disable:4996)
// #pragma comment(lib, "Netapi32.lib")
// #pragma comment(lib, "Advapi32.lib")
#pragma comment(lib, "Ws2_32.lib")
#define strncasecmp strnicmp
#define strcasecmp stricmp
typedef unsigned _int32 uint32;
#else
typedef uint32_t uint32;
#endif

#ifndef WIN32
#define GENERIC_WRITE 0x40000000
#define GENERIC_READ 0x80000000
#endif

#ifdef WIN32
#define BULLET "x"
#else
#define BULLET "\u2022"
#endif

long fileSizeLow, fileSizeHigh, blockSize;
void fillRandom (unsigned char *buffer, long len);
void fillConstant (unsigned char *buffer, unsigned char cData, long len);
void lsrand (unsigned int _seed);
long lrand (long maxRet);
double getTime (void);
char *lpretty (long long value);
char *llpretty (long long value);
char *pretty (char *buffer);
double reduceValue (char *scale, double r);
long getNumber (char *buf);
long long getNumberl (char *buf);
int getLongArg (char **argv, int x, long long *result);
int getLongArg (char **argv, int x, long *result);
long long strtoll (char *buf, char **eos, int base);
long long getFilesize (char *filename);
long compareBuffers (char *buffer, char *buffer1, long length);
long getDirCount (const char *base);
long getFileCount (const char *base);
bool openFile (long openType, char *filename);
void closeFile (void);
long writeFile (unsigned char *buffer, long count);
long readFile (unsigned char *buffer, long bytes);
void backup (int count);
int compareDoubles (const void *a, const void *b);
void usleep (double sec);
#ifdef WIN32
void dispError (FILE *fout);
#endif

char pBuffer[512];
#ifdef WIN32
HANDLE hFile, hFile1;
#endif
FILE *fFile, *sout;
bool useHandles, appendMode, flushIO, writeThru, doubleOpen;
long long _fileOffset, _writeErrors, _readErrors;
#define BLOCKPOOLMAX 8192
int blockPool;
char *pPool[BLOCKPOOLMAX];

enum cmType { CMNONE = 0, CMCONTINUE, CMLOOP, CMBAIL };


/* ------------------------------------------------------------
 * Function:	main
 * ------------------------------------------------------------ */
int main (int argc, char **argv)
{
char *filename = new char[8192], *statsName = new char[1024], *buffer,
     *buffer1, scale[16], cData, *ptr, *trigger = new char[1024];
long long totalBytesWritten, totalBytesRead, autoStop, errors;
double sTime, eTime, lTime, upsTime, fileSTime, fileETime, s, s1, rate;
long dirCount, fileCount, d, f, fileSizeLow, fileSizeHigh, fileSize,
     totalFiles, bytesLeft, count, bytesWritten, bytesRead,
     fileOffset, blocks, blockSize, _blockSize, lresp, retryCount;
unsigned int seed = 0;
int x, h, m, h1, m1;
bool verbose, randomData = true, variableBlocks = false,
     readMode = false, dualMode = false, dualState = false,
     cleanMode = false, andExit = false, optimalMode = false,
     continueIO = false, loop = false, lockBlock = false,
     unlockBlock = false;
cmType checkMissing;

blockPool = 0;
for (x = 0; x < BLOCKPOOLMAX; x++) pPool[x] = (char *)NULL;

*statsName = 0;
fileSizeLow = 1; fileSizeHigh = 1048576;
dirCount = 0; fileCount = 0;
blockSize = 16384;
_blockSize = 1024;			// For optimalMode
errors = 0;
cData = 0;
autoStop = 0;
useHandles = false; appendMode = false; flushIO = false;
writeThru = false;
sout = (FILE *)NULL;
*trigger = 0;
checkMissing = CMNONE;

for (x = 1; x < argc; x++) {
	if (argv[x][0] == '-') {
		if ((strcasecmp (argv[x], "--autoterm") == 0) || (strncmp (argv[x], "-a", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) autoStop = getNumberl (argv[++x]);
			else autoStop = getNumberl (&argv[x][2]);
			}
		else if (strcasecmp (argv[x], "--noAutoTerm") == 0) {
			autoStop = 0;
			}
		else if ((strcasecmp (argv[x], "--dirs") == 0) || (strncmp (argv[x], "-d", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) dirCount = getNumber (argv[++x]);
			else dirCount = getNumber (&argv[x][2]);
			}
		else if ((strcasecmp (argv[x], "--files") == 0) || (strncmp (argv[x], "-f", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) fileCount = getNumber (argv[++x]);
			else fileCount = getNumber (&argv[x][2]);
			}
		else if ((strcasecmp (argv[x], "--minFileSize") == 0) || (strncmp (argv[x], "-sl", 3) == 0)) {
			if (argv[x][1] == '-' || argv[x][3] == 0) fileSizeLow = getNumber (argv[++x]);
			else fileSizeLow = getNumber (&argv[x][3]);
			}
		else if ((strcasecmp (argv[x], "--maxFileSize") == 0) || (strncmp (argv[x], "-sh", 3) == 0)) {
			if (argv[x][1] == '-' || argv[x][3] == 0) fileSizeHigh = getNumber (argv[++x]);
			else fileSizeHigh = getNumber (&argv[x][3]);
			}
		else if ((strcasecmp (argv[x], "--fileSize") == 0) || (strncmp (argv[x], "-s", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) fileSizeLow = getNumber (argv[++x]);
			else fileSizeLow = getNumber (&argv[x][2]);
			fileSizeHigh = fileSizeLow;
			}
		else if ((strcasecmp (argv[x], "--stats") == 0) || (strncmp (argv[x], "-S", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) strcpy (statsName, argv[++x]);
			else strcpy (statsName, &argv[x][2]);
			}
		else if (strcasecmp (argv[x], "--noStats") == 0) {
			*statsName = 0;
			}
		else if ((strcasecmp (argv[x], "--blockSize") == 0) || (strncmp (argv[x], "-b", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) blockSize = getNumber (argv[++x]);
			else blockSize = getNumber (&argv[x][2]);
			// If they've manually set in, save it in case they're doing --optimal
			_blockSize = blockSize;
			}
		else if ((strcasecmp (argv[x], "--verbose") == 0) || (strcmp (argv[x], "-v") == 0)) {
			verbose = true;
			}
		else if (strcasecmp (argv[x], "--noVerbose") == 0) {
			verbose = false;
			}
		else if ((strcasecmp (argv[x], "--constant") == 0) || (strncmp (argv[x], "-c", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) cData = (unsigned char)atoi (argv[++x]);
			else cData = (unsigned char)atoi (&argv[x][2]);
			randomData = false;
			}
		else if ((strcasecmp (argv[x], "--seed") == 0) || (strncmp (argv[x], "-r", 2) == 0)) {
			if (argv[x][1] == '-' || argv[x][2] == 0) seed = atoi (argv[++x]);
			else seed = atoi (&argv[x][2]);
			}
		else if ((strcasecmp (argv[x], "--variableBlocks") == 0) || (strcmp (argv[x], "-V") == 0)) {
			variableBlocks = true;
			}
		else if (strcasecmp (argv[x], "--noVariableBlocks") == 0) {
			variableBlocks = false;
			}
		else if ((strcasecmp (argv[x], "--flush") == 0) || (strcmp (argv[x], "-F") == 0)) {
			flushIO = true;
			}
		else if (strcasecmp (argv[x], "--noFlush") == 0) {
			flushIO = false;
			}
		else if ((strcasecmp (argv[x], "--read") == 0) || (strcmp (argv[x], "-R") == 0)) {
			readMode = true;
			}
		else if ((strcasecmp (argv[x], "--write") == 0) || (strcmp (argv[x], "-W") == 0)) {
			readMode = false;
			}
		else if (strcasecmp (argv[x], "--writeThru") == 0) {
			writeThru = true;
			}
		else if (strcasecmp (argv[x], "--noWriteThru") == 0) {
			writeThru = false;
			}
		else if (strcasecmp (argv[x], "--trigger") == 0) {
			strcpy (trigger, argv[++x]);
			}
		else if (strcasecmp (argv[x], "--noTrigger") == 0) {
			*trigger = 0;
			}
#ifdef WIN32
		else if (strcasecmp (argv[x], "--useHandles") == 0) {
			useHandles = true;
			}
		else if (strcasecmp (argv[x], "--NoUseHandles") == 0) {
			useHandles = false;
			}
#endif
		else if (strcasecmp (argv[x], "--dualMode") == 0) {
			dualMode = true;
			}
		else if (strcasecmp (argv[x], "--noDualMode") == 0) {
			dualMode = false;
			}
		else if (strcasecmp (argv[x], "--clean") == 0) {		// Only specified on the command-line
			cleanMode = true;
			andExit = true;
			}
		else if (strcasecmp (argv[x], "--noClean") == 0) {
			cleanMode = false;
			andExit = false;
			}
		else if (strcasecmp (argv[x], "--optimal") == 0) {		// Optimal block size/options test
			optimalMode = true;
			}
		else if (strcasecmp (argv[x], "--noOptimal") == 0) {
			optimalMode = false;
			}
		else if (strcasecmp (argv[x], "--doubleOpen") == 0) {
			doubleOpen = true;
			}
		else if (strcasecmp (argv[x], "--noDoubleOpen") == 0) {
			doubleOpen = false;
			}
		else if (strcasecmp (argv[x], "--continue") == 0) {
			continueIO = true;
			}
		else if (strcasecmp (argv[x], "--noContinue") == 0) {
			continueIO = false;
			}
		else if (strcasecmp (argv[x], "--loop") == 0) {
			loop = true;
			}
		else if (strcasecmp (argv[x], "--noloop") == 0) {
			loop = false;
			}
		else if (strcasecmp (argv[x], "--checkMissing") == 0) {
			x++;
			if (strcasecmp (argv[x], "none") == 0) checkMissing = CMNONE;
			else if (strcasecmp (argv[x], "continue") == 0) checkMissing = CMCONTINUE;
			else if (strcasecmp (argv[x], "loop") == 0) checkMissing = CMLOOP;
			else if (strcasecmp (argv[x], "bail") == 0) checkMissing = CMBAIL;
			}
		else if (strcasecmp (argv[x], "--noCheckMissing") == 0) {
			checkMissing = CMNONE;
			}
		else if ((strcasecmp (argv[x], "--lockBlock") == 0) || (strcmp (argv[x], "-lb") == 0)) {
			lockBlock = true;
			}
		else if (strcasecmp (argv[x], "--noLockBlock") == 0) {
			lockBlock = false;
			}
		else if ((strcasecmp (argv[x], "--unlockBlock") == 0) || (strcmp (argv[x], "-ub") == 0)) {
			unlockBlock = true;
			}
		else if (strcasecmp (argv[x], "--noUnlockBlock") == 0) {
			unlockBlock = false;
			}
		else if (strcasecmp (argv[x], "--blockPool") == 0) {
			int count = atoi (argv[++x]);
			if (count > BLOCKPOOLMAX) fprintf (stderr, "ERROR: Maximum block pool count is %d\n", BLOCKPOOLMAX);
			else blockPool = count;
			}
		else if (strcasecmp (argv[x], "--noBlockPool") == 0) {
			blockPool = 0;
			}
		else if (strcasecmp (argv[x], "--writeConfig") == 0) {	// Write configuration
			FILE *fout;
			if ((fout = fopen (argv[++x], "wb")) == (FILE *)NULL) {
				fprintf (stderr, "ERROR: Cannot open config file %s for writing.\n", argv[x]);
				exit (1);
				}
			fprintf (fout, "# makeSet1 configuration file\n");
			fprintf (fout, "\n");
			fprintf (fout, "autoterm %ld\n", autoStop);
			fprintf (fout, "dirs %ld\n", dirCount);
			fprintf (fout, "files %ld\n", fileCount);
			if (fileSizeLow == fileSizeHigh) {
				fprintf (fout, "fileSize %ld\n", fileSizeLow);
				}
			else {
				fprintf (fout, "minFileSize %ld\n", fileSizeLow);
				fprintf (fout, "maxFileSize %ld\n", fileSizeHigh);
				}
			fprintf (fout, "blockSize %ld\n", blockSize);
			fprintf (fout, "seed %ld\n", seed);
			if (*statsName != 0) fprintf (fout, "stats %s\n", statsName);
			if (!randomData) fprintf (fout, "constant %d\n", (int)cData);
			if (verbose) fprintf (fout, "verbose\n");
			if (variableBlocks) fprintf (fout, "variableBlocks\n");
			if (flushIO) fprintf (fout, "flush\n");
			if (readMode) fprintf (fout, "read\n");
			else fprintf (fout, "write\n");
			if (writeThru) fprintf (fout, "writeThru\n");
			if (dualMode) fprintf (fout, "dualMode\n");
#ifdef WIN32
			if (useHandles) fprintf (fout, "useHandles\n");
			if (doubleOpen) fprintf (fout, "doubleOpen\n");
#endif
			if (cleanMode) fprintf (fout, "clean\n");
			if (*trigger != 0) fprintf (fout, "trigger %s\n", trigger);
			if (continueIO) fprintf (fout, "continue\n");
			if (loop) fprintf (fout, "loop\n");
			if (checkMissing != CMNONE) {
				switch (checkMissing) {
					case CMCONTINUE:	fprintf (fout, "checkMissing continue\n"); break;
					case CMLOOP:		fprintf (fout, "checkMissing loop\n"); break;
					case CMBAIL:		fprintf (fout, "checkMissing bail\n"); break;
					}
				}
			if (lockBlock) fprintf (fout, "lockBlock\n");
			if (unlockBlock) fprintf (fout, "unlockBlock\n");
			if (blockPool > 0) fprintf (fout, "blockPool %d\n", blockPool);
			fclose (fout);
			}
		else {
			if ((strcmp (argv[x], "--help") == 0) || (strcmp (argv[x], "-h") == 0) || (strcmp (argv[x], "-?") == 0)) {
				fprintf (stderr, "usage: makeSet1 options [configFile]\n");
				fprintf (stderr, "       -a value | --autoterm value (--noAutoTerm)\n");
				fprintf (stderr, "          Specify auto terminate byte count (read/write)\n");
				fprintf (stderr, "       -b value | --blockSize value\n");
				fprintf (stderr, "          Block size for I/O operations (default=16KB) (read/write)\n");
				fprintf (stderr, "       -c value | --constant value\n");
				fprintf (stderr, "          Fill buffer with constant (decimal argument) (read/write)\n");
				fprintf (stderr, "       -d value | --dirs value\n");
				fprintf (stderr, "          Specify directory count (read/write).\n");
				fprintf (stderr, "       -F | --flush (--noFlush)\n");
				fprintf (stderr, "          Flush write operations (write)\n");
				fprintf (stderr, "       -f value | --files value\n");
				fprintf (stderr, "          File count per directory (read/write)\n");
				fprintf (stderr, "       -lb | --lockBlock (--noLockBlock)\n");
				fprintf (stderr, "          Lock block prior to I/O\n");
				fprintf (stderr, "       -R | --read\n");
				fprintf (stderr, "          Read dataset (default is write)\n");
				fprintf (stderr, "       -r seed | --seed seed\n");
				fprintf (stderr, "          RNG seed (default=0) (read/write)\n");
				fprintf (stderr, "       -S pathname | --stats pathname (--noStats)\n");
				fprintf (stderr, "          Statistics filename (errors will cause a stop without a stats file) (read/write)\n");
				fprintf (stderr, "       -s value | --fileSize value\n");
				fprintf (stderr, "          Specify filesize to create (write only)\n");
				fprintf (stderr, "       -sl value | --minFileSize value\n");
				fprintf (stderr, "          Minimum filesize (default=1B) (write only)\n");
				fprintf (stderr, "       -sh value | --maxFileSize value\n");
				fprintf (stderr, "          Maximum filesize (default=1MB) (write only)\n");
				fprintf (stderr, "       -ub | --unlockBlock (--noUnlockBlock)\n");
				fprintf (stderr, "          Unlock block after I/O\n");
				fprintf (stderr, "       -V | --variableBlocks (--noVariableBlocks)\n");
				fprintf (stderr, "          Variable block sizes (range 1..blockSize..2*blockSize) (read/write)\n");
				fprintf (stderr, "       -v | --verbose (--noVerbose)\n");
				fprintf (stderr, "          Verbose output (read/write)\n");
				fprintf (stderr, "       -W | --write\n");
				fprintf (stderr, "          Write dataset\n");
				fprintf (stderr, "       --blockPool num (--noBlockPool)\n");
				fprintf (stderr, "         Pre-generate pool of blocks to derive data from\n");
				fprintf (stderr, "         If constant, each block is +1, otherwise each block is uniquely random\n");

				fprintf (stderr, "       --checkMissing continue|loop|bail (--noCheckMissing)\n");
				fprintf (stderr, "         If testfile is missing, continue onto the next file, start the test over,\n");
				fprintf (stderr, "         or bail out of the test. (read only)\n");
				fprintf (stderr, "       --clean (--noClean)\n");
				fprintf (stderr, "         Remove dirs/files before running\n");
				fprintf (stderr, "         If specified on the commandline, will exit after clean\n");
				fprintf (stderr, "       --continue (--noContinue)\n");
				fprintf (stderr, "         Continue where test stopped (write only)\n");
				#ifdef WIN32
				fprintf (stderr, "       --doubleOpen (--noDoubleOpen)\n");
				fprintf (stderr, "         Double open the file (break down to oplock none)\n");
				fprintf (stderr, "         Only useful with useHandles enabled\n");
				#endif
				fprintf (stderr, "       --dualMode (--noDualMode)\n");
				fprintf (stderr, "         Alternate blocks between constant & random data\n");
				fprintf (stderr, "       --loop (--noLoop)\n");
				fprintf (stderr, "         Wrap test with a loop - useful to continually read from a dataset\n");
				fprintf (stderr, "       --optimal (--noOptimal)\n");
				fprintf (stderr, "         Generate I/O until optimal blocksize/options are found\n");
				fprintf (stderr, "       --trigger pathname (--noTrigger)\n");
				fprintf (stderr, "         Touch pathname to synchronously start multiple instances.\n");
				fprintf (stderr, "         Once running, remove pathname to stop all instances.\n");
				#ifdef WIN32
				fprintf (stderr, "       --useHandles (--noUseHandles)\n");
				fprintf (stderr, "         Use handle-based I/O (windows-only)\n");
				#endif
				fprintf (stderr, "       --writeConfig pathname\n");
				fprintf (stderr, "          Write config file based upon current parameters\n");
				fprintf (stderr, "       --writeThru (--noWriteThru)\n");
				fprintf (stderr, "         Attempt to throw file into writethru/nobuffer mode\n");

				fprintf (stderr, "\n");
				fprintf (stderr, "%s Counts can be specified in canonical form, or with units.\n", BULLET);
				fprintf (stderr, "  1KB is 1024 bytes.  1048576 = 1MB.\n");
				fprintf (stderr, "  KB, MB, GB, TB accepted.\n");
				fprintf (stderr, "%s Options after configuration file override settings in configuration file.\n", BULLET);
				fprintf (stderr, "%s Use the --noOptions to turn off parameters set in the config file.\n", BULLET);
				fprintf (stderr, "%s To send locks (-lb) across wire, double open (--doubleOpen) must be used\n", BULLET);
				fprintf (stderr, "\n");
				}
			else {
				fprintf (stderr, "ERROR: unknown option %s\n", argv[x]);
				fprintf (stderr, "       Use --help for usage information\n");
				}
			exit (1);
			}
		}
	else {						// Read configuration file
		FILE *fin;
		if ((fin = fopen (argv[x], "rb")) == (FILE *)NULL) {
			fprintf (stderr, "ERROR: Cannot open configuration file %s for reading.\n", argv[x]);
			exit (1);
			}
		buffer = new char[16384];
		while (1) {
			memset (buffer, 0, 8192);
			fgets (buffer, 8192, fin);
			if (feof (fin)) break;
			if ((ptr = strchr (buffer, 0x0d)) != (char *)NULL) *ptr = 0;
			if ((ptr = strchr (buffer, 0x0a)) != (char *)NULL) *ptr = 0;
			if (*buffer == '#') continue;
			if (strlen (buffer) < 2) continue;
			if ((ptr = strchr (buffer, ' ')) != (char *)NULL) {
				while (1) {
					if (*ptr != ' ' && *ptr != 0x09) break;
					ptr++;
					}
				}
			if (strncasecmp (buffer, "autoterm ", 9) == 0) autoStop = getNumberl (ptr);
			else if (strncasecmp (buffer, "dirs ", 5) == 0) dirCount = getNumber (ptr);
			else if (strncasecmp (buffer, "files ", 6) == 0) fileCount = getNumber (ptr);
			else if (strncasecmp (buffer, "filesize ", 9) == 0) { fileSizeLow = getNumber (ptr); fileSizeHigh = fileSizeLow; }
			else if (strncasecmp (buffer, "minFileSize ", 12) == 0) fileSizeLow = getNumber (ptr);
			else if (strncasecmp (buffer, "maxFileSize ", 12) == 0) fileSizeHigh = getNumber (ptr);
			else if (strncasecmp (buffer, "blockSize ", 10) == 0) blockSize = getNumber (ptr);
			else if (strncasecmp (buffer, "seed ", 5) == 0) seed = atoi (ptr);
			else if (strncasecmp (buffer, "stats ", 6) == 0) strcpy (statsName, ptr);
			else if (strncasecmp (buffer, "constant ", 9) == 0) {
				cData = (unsigned char)atoi (ptr);
				randomData = false;
				}
			else if (strcasecmp (buffer, "verbose") == 0) verbose = true;
			else if (strcasecmp (buffer, "variableBlocks") == 0) variableBlocks = true;
			else if (strcasecmp (buffer, "flush") == 0) flushIO = true;
			else if (strcasecmp (buffer, "read") == 0) readMode = true;
			else if (strcasecmp (buffer, "write") == 0) readMode = false;
			else if (strcasecmp (buffer, "writeThru") == 0) writeThru = true;
			else if (strcasecmp (buffer, "dualMode") == 0) dualMode = true;
#ifdef WIN32
			else if (strcasecmp (buffer, "useHandles") == 0) useHandles = true;
			else if (strcasecmp (buffer, "doubleOpen") == 0) doubleOpen = true;
#endif
			else if (strcasecmp (buffer, "clean") == 0) cleanMode = true;
			else if (strncasecmp (buffer, "trigger ", 8) == 0) strcpy (trigger, ptr);
			else if (strcasecmp (buffer, "continue") == 0) continueIO = true;
			else if (strcasecmp (buffer, "loop") == 0) loop = true;
			else if (strncasecmp (buffer, "blockPool", 9) == 0) blockPool = (int)getNumber (ptr);
			else if (strncasecmp (buffer, "checkMissing ", 13) == 0) {
				if (strcasecmp (ptr, "none") == 0) checkMissing = CMNONE;
				else if (strcasecmp (ptr, "continue") == 0) checkMissing = CMCONTINUE;
				else if (strcasecmp (ptr, "loop") == 0) checkMissing = CMLOOP;
				else if (strcasecmp (ptr, "bail") == 0) checkMissing = CMBAIL;
				}
			}
		fclose (fin);
		delete[] buffer;
		}	// read configuration
	}	// for x

lsrand (seed);		// Seed the RNG
if (readMode || cleanMode) {
	if (dirCount == 0) dirCount = getDirCount (".");
	if (fileCount == 0) fileCount = getFileCount (".");
	}

if ((dirCount < 1 || fileCount < 1) && !optimalMode) {
	fprintf (stderr, "Error: nothing to do, dirs=%ld, files per dir=%ld\n", dirCount, fileCount);
	exit (1);
	}

if (*trigger != 0) {
	sTime = getTime ();
	while (1) {
		s = getTime () - sTime;
		m = (int)(s / 60);
		s = s - (m * 60);
		if (verbose) fprintf (stdout, "\rWaiting for trigger: %02d:%02.0f ", m, s);
		if (access (trigger, 0) == 0) break;			// Trigger found, continue
		usleep (0.1);
		}
	if (verbose) fprintf (stdout, "\rWaiting for trigger: %02d:%02.0f (triggered)\n", m, s);
	}

// Clean up any mess left behind
if (cleanMode) {
	if (verbose) fprintf (stdout, "----- Cleaning directory structure -----\n\n");
	for (f = 0; f < fileCount; f++) {
		for (d = 0; d < dirCount; d++) {
			sprintf (filename, "subdir%06ld/file%06ld.dat", d, f);
			unlink (filename);
			}
		}
	for (d = 0; d < dirCount; d++) {
		sprintf (filename, "subdir%06ld", d);
		rmdir (filename);
		}
	if (andExit) exit (0);
	}

sout = (FILE *)NULL;
if (*statsName != 0) {
	if ((sout = fopen (statsName, "wb")) == (FILE *)NULL) {
		fprintf (stderr, "ERROR: Cannot open stats file %s for writing.\n", statsName);
		}
	}

// ----- Optimal Test -----
if (optimalMode) {
	if (verbose) fprintf (stdout, "----- Optimal Blocksize Test -----\n\n");
	struct _stats {
		long blockSize, ops;
		bool useHandles, flushIO, writeThru;
		double minOpTime, maxOpTime, totalTime, openTime, closeTime;
		double nMinOpTime, nMaxOpTime, nAvgTime;		// The 90% values
		} stats[1000], tmpStat;
	int maxCase = 0, y;
	long timesSize = fileSizeHigh / 1024, maxTimes, lx;
	double *times = new double[timesSize], dx, dx1;
	bool bx;
	for (x = 0; x < 1000; x++) {
		stats[x].blockSize = 0;
		}
	// Fill in the stats structure and let it drive the I/O loop
	useHandles = false; flushIO = false; writeThru = false;
	x = 0;
	for (blockSize = _blockSize; blockSize <= 256 * 1024; blockSize = blockSize * 2) {
		while (1) {
			stats[x].blockSize = blockSize;
			stats[x].useHandles = useHandles;
			stats[x].flushIO = flushIO;
			stats[x].writeThru = writeThru;
			stats[x].minOpTime = 99999.9;
			stats[x].maxOpTime = 0.0;
			stats[x].totalTime = 0.0;
			stats[x].ops = 0;
			maxCase = x;
			x++;
			writeThru = !writeThru;
			if (writeThru == false) {
				flushIO = !flushIO;
				if (flushIO == false) {
#ifdef WIN32
					useHandles = !useHandles;
					if (useHandles == false) {
						break;
						}
#else
					break;
#endif
					}
				}
			}	// while
		}		// blocksize

	// I/O loop - create files and update the stats table
	double startTime, opStartTime, diffTime, now;
	buffer = new char[1024 * 1024];
	strcpy (filename, "optimal.dat");
	fprintf (stdout, "Running test cases\n");
	fprintf (stdout, "\n");

	fprintf (stdout, "        Block                           Open                           I/O      Close               90%%      90%% I/O   Open\n");
	fprintf (stdout, "Case     Size     I/O Ops    Time       Time        I/O Range      Average       Time         I/O Range      Average  Flags\n");
	fprintf (stdout, "-----  ------ ----------- ------- ---------- ---------------- ------------ ----------  ---------------- ------------  -----\n");
//  fprintf (stdout, "00/39:   16KB Ops:  32000 00:23.8 O:0.01585s 0.00035-0.85480s Avg:0.00074s C:0.00054s (0.00040-0.00074s Avg:0.00053s) F:HFW\n");
	for (x = 0; ; x++) {
		if (stats[x].blockSize == 0) break;
		blockSize = stats[x].blockSize;
		useHandles= stats[x].useHandles;
		flushIO = stats[x].flushIO;
		writeThru = stats[x].writeThru;
		stats[x].maxOpTime = 0.0;
		stats[x].minOpTime = 999999.9;
		startTime = getTime ();
		opStartTime = getTime ();
		if (openFile (GENERIC_WRITE, filename)) {
			stats[x].openTime = getTime () - opStartTime;
			bytesLeft = fileSizeHigh;
			upsTime = getTime ();
			while (1) {
				if (bytesLeft < blockSize) count = bytesLeft; else count = blockSize;
				if (dualMode) {
					if (dualState) fillRandom ((unsigned char *)buffer, count);
					else fillConstant ((unsigned char *)buffer, cData, count);
					dualState = !dualState;
					}
				else {
					if (randomData) fillRandom ((unsigned char *)buffer, count);
					else fillConstant ((unsigned char *)buffer, cData, count);
					}
				opStartTime = getTime ();
				bytesWritten = writeFile ((unsigned char *)buffer, count);
				now = getTime ();
				diffTime = now - opStartTime;
				times[stats[x].ops] = diffTime;				// Record this time
				if (diffTime < stats[x].minOpTime) stats[x].minOpTime = diffTime;
				if (diffTime > stats[x].maxOpTime) stats[x].maxOpTime = diffTime;
				stats[x].ops++;
				bytesLeft -= bytesWritten;
				stats[x].totalTime = now - startTime;
				if (verbose) {
					if ((now - upsTime) > 0.1) {
						s = now - startTime;
						m = (int)(s / 60.0);
						s = s - (m * 60.0);
						fprintf (stdout, "\r%02d/%02d: %4ldKB Ops:%7ld %02d:%04.1f O:%.5fs %.5f-%.5fs Avg:%.5fs C:         (                 Avg:        ) F:%c%c%c  ",
						  x, maxCase, blockSize / 1024, stats[x].ops, m, s, stats[x].openTime, stats[x].minOpTime, stats[x].maxOpTime, stats[x].totalTime / stats[x].ops,
						  (stats[x].useHandles) ? 'H' : 'h', (stats[x].flushIO) ? 'F' : 'f', (stats[x].writeThru) ? 'W' : 'w');
						upsTime = now;
						}
					}
				if (bytesLeft <= 0) break;
				if (*trigger != 0) if (access (trigger, 0) != 0) break;
				}	// while
			opStartTime = getTime ();
			closeFile ();
			}
		now = getTime ();
		stats[x].closeTime = now - opStartTime;
		stats[x].totalTime = now - startTime;
		unlink (filename);				// Remove it for the next pass
		// Calc the 90% values
		maxTimes = fileSizeHigh / blockSize;
		qsort (times, maxTimes, sizeof (double), compareDoubles);
		dx = maxTimes * .05;							// 5% gap on bottom/top of list
		stats[x].nMinOpTime = times[(long)dx];
		stats[x].nMaxOpTime = times[(long)(maxTimes - dx)];
		stats[x].nAvgTime = 0.0;
		for (y = (int)dx; y <= (int)(maxTimes - dx); y++) {
			stats[x].nAvgTime += times[y];		// It's total right now
			}
		stats[x].nAvgTime = stats[x].nAvgTime / (maxTimes * .9);
		if (verbose) {
			s = now - startTime;
			m = (int)(s / 60.0);
			s = s - (m * 60.0);
			fprintf (stdout, "\r%02d/%02d: %4ldKB Ops:%7ld %02d:%04.1f O:%.5fs %.5f-%.5fs Avg:%.5fs C:%.5fs (%.5f-%.5fs Avg:%.5fs) F:%c%c%c          \n",
			  x, maxCase, blockSize / 1024, stats[x].ops, m, s, stats[x].openTime, stats[x].minOpTime, stats[x].maxOpTime, stats[x].totalTime / stats[x].ops, stats[x].closeTime,
			  stats[x].nMinOpTime, stats[x].nMaxOpTime, stats[x].nAvgTime,
			  (stats[x].useHandles) ? 'H' : 'h', (stats[x].flushIO) ? 'F' : 'f', (stats[x].writeThru) ? 'W' : 'w');
			fflush (stdout);
			}
		if (*trigger != 0) if (access (trigger, 0) != 0) break;
		}	// for x

	// NOTE: Read case here...

	// ----- Dump out stats array -----
	fprintf ((sout == (FILE *)NULL) ? stdout : sout, "\n");
	fprintf ((sout == (FILE *)NULL) ? stdout : sout, "I/O rates by blocksize\n");
	for (x = 0; ; x++) {
		if (stats[x].blockSize == 0) break;
		dx = reduceValue (scale, stats[x].blockSize / (stats[x].totalTime / stats[x].ops));
		fprintf ((sout == (FILE *)NULL) ? stdout : sout, "  %2d. %4ldKB %c%c%c Rate:%7.2f%s", x, stats[x].blockSize / 1024,
		  (stats[x].useHandles) ? 'H' : 'h', (stats[x].flushIO) ? 'F' : 'f', (stats[x].writeThru) ? 'W' : 'w',
		  dx, scale);
		dx = reduceValue (scale, stats[x].blockSize / stats[x].nAvgTime);
		fprintf ((sout == (FILE *)NULL) ? stdout : sout, ", 90%%:%7.2f%s\n", dx, scale);
		}

	// Sort on average
	for (x = 0; ; x++) {
		if (stats[x+1].blockSize == 0) break;
		dx = stats[x].blockSize / (stats[x].totalTime / stats[x].ops);
		dx1 = stats[x+1].blockSize / (stats[x+1].totalTime / stats[x+1].ops);
		if (dx < dx1) {
			memcpy ((void *)&tmpStat, (void *)&stats[x], sizeof (struct _stats));
			memcpy ((void *)&stats[x], (void *)&stats[x+1], sizeof (struct _stats));
			memcpy ((void *)&stats[x+1], (void *)&tmpStat, sizeof (struct _stats));
			if (x < 2) x = -1; else x -= 2;
			}
		}
	fprintf ((sout == (FILE *)NULL) ? stdout : sout, "\n");
	fprintf ((sout == (FILE *)NULL) ? stdout : sout, "I/O rates by throughput (overall average)\n");
	for (x = 0; ; x++) {
		if (stats[x].blockSize == 0) break;
		dx = reduceValue (scale, stats[x].blockSize / (stats[x].totalTime / stats[x].ops));
		fprintf ((sout == (FILE *)NULL) ? stdout : sout, "  %2d. %4ldKB %c%c%c Rate:%7.2f%s", x, stats[x].blockSize / 1024,
		  (stats[x].useHandles) ? 'H' : 'h', (stats[x].flushIO) ? 'F' : 'f', (stats[x].writeThru) ? 'W' : 'w',
		  dx, scale);
		dx = reduceValue (scale, stats[x].blockSize / stats[x].nAvgTime);
		fprintf ((sout == (FILE *)NULL) ? stdout : sout, ", 90%%:%7.2f%s\n", dx, scale);
		}

	// Sort on average
	for (x = 0; ; x++) {
		if (stats[x+1].blockSize == 0) break;
		dx = stats[x].blockSize / stats[x].nAvgTime;
		dx1 = stats[x+1].blockSize / stats[x+1].nAvgTime;
		if (dx < dx1) {
			memcpy ((void *)&tmpStat, (void *)&stats[x], sizeof (struct _stats));
			memcpy ((void *)&stats[x], (void *)&stats[x+1], sizeof (struct _stats));
			memcpy ((void *)&stats[x+1], (void *)&tmpStat, sizeof (struct _stats));
			if (x < 2) x = -1; else x -= 2;
			}
		}
	fprintf ((sout == (FILE *)NULL) ? stdout : sout, "\n");
	fprintf ((sout == (FILE *)NULL) ? stdout : sout, "I/O rates by throughput (90%% average)\n");
	for (x = 0; ; x++) {
		if (stats[x].blockSize == 0) break;
		dx = reduceValue (scale, stats[x].blockSize / (stats[x].totalTime / stats[x].ops));
		fprintf ((sout == (FILE *)NULL) ? stdout : sout, "  %2d. %4ldKB %c%c%c Rate:%7.2f%s", x, stats[x].blockSize / 1024,
		  (stats[x].useHandles) ? 'H' : 'h', (stats[x].flushIO) ? 'F' : 'f', (stats[x].writeThru) ? 'W' : 'w',
		  dx, scale);
		dx = reduceValue (scale, stats[x].blockSize / stats[x].nAvgTime);
		fprintf ((sout == (FILE *)NULL) ? stdout : sout, ", 90%%:%7.2f%s\n", dx, scale);
		}

	delete[] buffer;
	if (sout != (FILE *)NULL) fclose (sout);
	exit (0);
	} // optimal mode

_blockSize = blockSize;			// Save for variable block calc

if (variableBlocks) {
	buffer = new char[(blockSize * 2) + 1];
	if (readMode) buffer1 = new char[(blockSize * 2) + 1];
	}
else {
	buffer = new char[blockSize];
	if (readMode) buffer1 = new char[blockSize];
	}

// If we're using a blockPool, allocate and initialize it now
dualState = false;
if (blockPool > 0) {
	int bp = blockPool;
	int cDataSave = cData;
	blockPool = 0;				// Turn it off so the generators work
	for (x = 0; x < bp; x++) {
		if (variableBlocks) count = (blockSize * 2) + 1;
		else count = blockSize;
		pPool[x] = new char[count];
		if (dualMode) {
			if (dualState) fillRandom ((unsigned char *)pPool[x], count);
			else {
				fillConstant ((unsigned char *)pPool[x], cData, count);
				cData = (char)((int)cData + 1);
				}
			dualState = !dualState;
			}
		else {
			if (randomData) fillRandom ((unsigned char *)pPool[x], count);
			else {
				fillConstant ((unsigned char *)pPool[x], cData, count);
				cData = (char)((int)cData + 1);
				}
			}
		}
	blockPool = bp;
	cData = cDataSave;
	}
dualState = false;

totalFiles = 0;
totalBytesWritten = 0;
totalBytesRead = 0;

if (sout != (FILE *)NULL) {
	fprintf (sout, "Command line:");
	for (x = 0; x < argc; x++) fprintf (sout, " %s", argv[x]);
	fprintf (sout, "\n");
	fprintf (sout, "\n");
	fprintf (sout, "Filename, FileSize, BlockSize, Constant/Variable, Blocks, Time, Errors\n");
	fflush (sout);

	sTime = getTime ();
	if (randomData) fillRandom ((unsigned char *)buffer, blockSize);
	else fillConstant ((unsigned char *)buffer, cData, blockSize);
	eTime = getTime ();
	fprintf (sout, "*** Buffer fill overhead: %.5fs\n", eTime - sTime);
	lsrand (seed);		// Seed the RNG
	}

// ----- Read Mode -----
if (readMode) {
	if (verbose) fprintf (stdout, "----- Read Test -----\n\n");
	sTime = getTime ();
	upsTime = sTime;
	do {					// loop test
		for (f = 0; f < fileCount; f++) {
			for (d = 0; d < dirCount; d++) {
				sprintf (filename, "subdir%06ld/file%06ld.dat", d, f);
				fileSTime = getTime ();
				// If this is set, we have special handling if the file isn't there
				if (checkMissing != CMNONE) {
					if (access (filename, 0) != 0) {
						if (checkMissing == CMCONTINUE) {		// Just move along to next file
							continue;
							}
						else if (checkMissing == CMLOOP) {		// Set to max, turn on loop, and continue
							f = fileCount; d = dirCount;
							loop = true;
							continue;
							}
						else if (checkMissing == CMBAIL) {		// Set to max, turn off loop, and continue
							f = fileCount; d = dirCount;
							loop = false;
							continue;
							}
						}
					}
				// Open the test file
				if (!openFile (GENERIC_READ, filename)) continue;						// Move along, maybe it'll fix itself
				bytesLeft = lrand (fileSizeHigh - fileSizeLow + 1) + fileSizeLow;		// Note: called to keep RNG in sync
				bytesLeft = (long)getFilesize (filename);
				fileSize = bytesLeft; fileOffset = 0L;
				blocks = 0;
				// Read loop
				while (1) {
					if (variableBlocks) blockSize = lrand (_blockSize * 2) + 1;
					if (bytesLeft < blockSize) count = bytesLeft; else count = blockSize;
					if (dualMode) {
						if (dualState) fillRandom ((unsigned char *)buffer, count);
						else fillConstant ((unsigned char *)buffer, cData, count);
						dualState = !dualState;
						}
					else {
						if (randomData) fillRandom ((unsigned char *)buffer, count);
						else fillConstant ((unsigned char *)buffer, cData, count);
						}
					errno = 0;
					bytesRead = readFile ((unsigned char *)buffer1, count);
					if ((lresp = compareBuffers (buffer, buffer1, bytesRead)) > 0) {
						errors += lresp;		// Each byte miscompare is an error
						if (sout != (FILE *)NULL)
							fprintf (sout, "%s, *** Data miscompare ***, Bytes Read:%ld, Error:%s (%d)\n", filename, bytesRead, strerror (errno), errno);
						}
					// NOTE: Should check for short read
					blocks++;
					totalBytesRead += bytesRead;
					bytesLeft -= bytesRead;
					fileOffset += bytesRead;
					if (bytesLeft <= 0) break;
					// If >1s, update the stats
					if ((getTime () - upsTime) > 1.0) {
						// Output rolling stats
						lTime = getTime () - sTime;
						h = 0; m = 0; s = lTime;
						while (s >= 60.0) {
							s -= 60.0;
							if (++m >= 60) { m -= 60; h++; }
							}
						fprintf (stdout, "\rFiles:%s ", lpretty (totalFiles));
						fprintf (stdout, "Total bytes:%s ", llpretty (totalBytesRead));
						fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h, m, s);
						rate = reduceValue (scale, totalBytesRead / lTime);
						fprintf (stdout, "Rate:%.2f%s  |  ", rate, scale);
						lTime = getTime () - fileSTime;
						h1 = 0; m1 = 0; s1 = lTime;
						while (s1 >= 60.0) {
							s1 -= 60.0;
							if (++m >= 60) { m -= 60; h++; }
							}
						fprintf (stdout, "Current:%s  ", lpretty (fileSize));
						fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h1, m1, s1);
						rate = reduceValue (scale, fileOffset / lTime);
						fprintf (stdout, "Rate:%.2f%s        ", rate, scale);
						fflush (stdout);
						upsTime = getTime ();
						}
					if ((autoStop > 0) && (totalBytesWritten >= autoStop)) break;
					if (*trigger != 0) if (access (trigger, 0) != 0) break;
					}			// File read loop
				closeFile ();
				totalFiles++;
				fileETime = getTime ();
				eTime = getTime ();
				lTime = eTime - sTime;

				// Write a stats record
				if (sout != (FILE *)NULL) {
					if (variableBlocks)
						fprintf (sout, "%s, %ld, %ld, variable, %ld, %f, %ld\n", filename, fileSize, fileSize / blocks, blocks, fileETime - fileSTime, errors);
					else
						fprintf (sout, "%s, %ld, %ld, constant, %ld, %f, %ld\n", filename, fileSize, _blockSize, blocks, fileETime - fileSTime, errors);
					fflush (sout);
					}
				// Output stats
				lTime = getTime () - sTime;
				h = 0; m = 0; s = lTime;
				while (s >= 60.0) {
					s -= 60.0;
					if (++m >= 60) { m -= 60; h++; }
					}
				fprintf (stdout, "\rFiles:%s ", lpretty (totalFiles));
				fprintf (stdout, "Total bytes:%s ", llpretty (totalBytesRead));
				fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h, m, s);
				rate = reduceValue (scale, totalBytesRead / lTime);
				fprintf (stdout, "Rate:%.2f%s  |  ", rate, scale);
				lTime = getTime () - fileSTime;
				h1 = 0; m1 = 0; s1 = lTime;
				while (s1 >= 60.0) {
					s1 -= 60.0;
					if (++m >= 60) { m -= 60; h++; }
					}
				fprintf (stdout, "Current:%s  ", lpretty (fileSize));
				fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h1, m1, s1);
				rate = reduceValue (scale, fileOffset / lTime);
				fprintf (stdout, "Rate:%.2f%s        ", rate, scale);
				fflush (stdout);

				if ((autoStop > 0) && (totalBytesWritten >= autoStop)) break;
				if (*trigger != 0) if (access (trigger, 0) != 0) break;
				}		// d = 0..dirCount
			if ((autoStop > 0) && (totalBytesRead >= autoStop)) break;
			if (*trigger != 0) if (access (trigger, 0) != 0) break;
			}	// f = 0..fileCount
		if ((autoStop > 0) && (totalBytesRead >= autoStop)) break;
		if (*trigger != 0) if (access (trigger, 0) != 0) break;
		fprintf (stdout, "\n");
		} while (loop);
	}	// readMode

// ----- Write Mode -----
else {
	if (verbose) fprintf (stdout, "----- Write Test -----\n\n");
	sTime = getTime ();
	upsTime = sTime;
	do {			// While loop
		for (f = 0; f < fileCount; f++) {
			for (d = 0; d < dirCount; d++) {
				// If first pass through, create the subdir
				if (f == 0) {
					sprintf (filename, "subdir%06ld", d);
					#ifdef WIN32
					_mkdir (filename);
					#else
					mkdir (filename, 0777);
					#endif
					}
				sprintf (filename, "subdir%06ld/file%06ld.dat", d, f);
				if (continueIO) {			// If we're continuing, only start writing once we find a missing file
					if (access (filename, 0) == 0) {
						fprintf (stdout, "\rContinuing d:%d/f:%d  ", d, f);
						continue;
						}
					continueIO = false;
					}
				fileSTime = getTime ();
				if (!openFile (GENERIC_WRITE, filename)) continue;						// Move along, maybe it'll fix itself.
				bytesLeft = lrand (fileSizeHigh - fileSizeLow + 1) + fileSizeLow;
				fileSize = bytesLeft; fileOffset = 0L;
				blocks = 0;
				retryCount = 0;

				while (1) {
					if (variableBlocks) blockSize = lrand (_blockSize * 2) + 1;
					if (bytesLeft < blockSize) count = bytesLeft; else count = blockSize;
					if (dualMode) {
						if (dualState) fillRandom ((unsigned char *)buffer, count);
						else fillConstant ((unsigned char *)buffer, cData, count);
						dualState = !dualState;
						}
					else {
						if (randomData) fillRandom ((unsigned char *)buffer, count);
						else fillConstant ((unsigned char *)buffer, cData, count);
						}
					errno = 0;
					bytesWritten = writeFile ((unsigned char *)buffer, count);
					// Log short-write error
					if (bytesWritten != count) {
						errors++;
						fprintf (sout, "%s, *** Short Write *** Offset:%ld, Block Size:%ld, Current Block Size:%ld, Bytes Written:%ld, Error:%s (%d)\n", filename, fileOffset, blockSize, count, bytesWritten, strerror (errno), errno);
						fflush (sout);
						if (++retryCount > 5) {
							fprintf (sout, "%s, *** Bailing *** Offset:%ld, Block Size:%ld, Current Block Size:%ld, Bytes Written:%ld\n", filename, fileOffset, blockSize, count, bytesWritten);
							break;		// Bail - problem with the file
							}
						}
					blocks++;
					totalBytesWritten += bytesWritten;
					bytesLeft -= bytesWritten;
					fileOffset += bytesWritten;
					if (bytesLeft <= 0) break;
					// If >1s, update the stats
					if ((getTime () - upsTime) > 1.0) {
						lTime = getTime () - sTime;
						h = 0; m = 0; s = lTime;
						while (s >= 60.0) {
							s -= 60.0;
							if (++m >= 60) { m -= 60; h++; }
							}
						fprintf (stdout, "\rFiles:%s ", lpretty (totalFiles));
						fprintf (stdout, "Total bytes:%s ", llpretty (totalBytesWritten));
						fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h, m, s);
						rate = reduceValue (scale, totalBytesWritten / lTime);
						fprintf (stdout, "Rate:%.2f%s  |  ", rate, scale);	
						lTime = getTime () - fileSTime;
						h1 = 0; m1 = 0; s1 = lTime;
						while (s1 >= 60.0) {
							s1 -= 60.0;
							if (++m >= 60) { m -= 60; h++; }
							}
						fprintf (stdout, "Current:%s  ", lpretty (fileSize));
						fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h1, m1, s1);
						rate = reduceValue (scale, fileOffset / lTime);
						fprintf (stdout, "Rate:%.2f%s        ", rate, scale);
						fflush (stdout);
						upsTime = getTime ();
						}
					if ((autoStop > 0) && (totalBytesWritten >= autoStop)) break;
					if (*trigger != 0) if (access (trigger, 0) != 0) break;
					}
				closeFile ();	
				totalFiles++;
				fileETime = getTime ();
				eTime = getTime ();
				lTime = eTime - sTime;

				// Write a stats record
				if (sout != (FILE *)NULL) {
					if (variableBlocks)
						fprintf (sout, "%s, %ld, %ld, variable, %ld, %f, %ld\n", filename, fileSize, fileSize / blocks, blocks, fileETime - fileSTime, errors);
					else
						fprintf (sout, "%s, %ld, %ld, constant, %ld, %f, %ld\n", filename, fileSize, _blockSize, blocks, fileETime - fileSTime, errors);
					fflush (sout);
					}

				// ----- Output stats
				lTime = getTime () - sTime;
				h = 0; m = 0; s = lTime;
				while (s >= 60.0) {
					s -= 60.0;
					if (++m >= 60) { m -= 60; h++; }
					}
				fprintf (stdout, "\rFiles:%s ", lpretty (totalFiles));
				fprintf (stdout, "Total bytes:%s ", llpretty (totalBytesWritten));
				fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h, m, s);
				rate = reduceValue (scale, totalBytesWritten / lTime);
				fprintf (stdout, "Rate:%.2f%s  |  ", rate, scale);
				lTime = getTime () - fileSTime;
				h1 = 0; m1 = 0; s1 = lTime;
				while (s1 >= 60.0) {
					s1 -= 60.0;
					if (++m >= 60) { m -= 60; h++; }
					}
				fprintf (stdout, "Current:%s  ", lpretty (fileSize));
				fprintf (stdout, "Elapsed:%02d:%02d:%04.1f ", h1, m1, s1);
				rate = reduceValue (scale, fileOffset / lTime);
				fprintf (stdout, "Rate:%.2f%s        ", rate, scale);
				fflush (stdout);

				if ((autoStop > 0) && (totalBytesWritten >= autoStop)) break;
				if (*trigger != 0) if (access (trigger, 0) != 0) break;
				}		// d = 0..dirCount
			if ((autoStop > 0) && (totalBytesWritten >= autoStop)) break;
			if (*trigger != 0) if (access (trigger, 0) != 0) break;
			}			// f = 0..fileCount
		if ((autoStop > 0) && (totalBytesWritten >= autoStop)) break;
		if (*trigger != 0) if (access (trigger, 0) != 0) break;
		} while (loop);
	}				// Write mode

if ((autoStop > 0) && (totalBytesWritten >= autoStop)) {
	if (sout != (FILE *)NULL) {
		fprintf (sout, "*** Early stop due to autostop.\n");
		}
	}
if (*trigger != 0) if (access (trigger, 0) != 0) {
	fprintf (sout, "*** Early stop due to trigger removal.\n");
	}
fprintf (stdout, "\n");
fflush (stdout);

if (sout != (FILE *)NULL) fclose (sout);
}


/* ------------------------------------------------------------
 * Function:    fillRandom
 * Description: Fill buffer with random data
 * ------------------------------------------------------------ */
void fillRandom (unsigned char *buffer, long len)
{
if (blockPool) {
	long l = lrand (blockPool);
	char *pPtr = pPool[l];
	memcpy (buffer, pPtr, len);
	}
else {
	for (long lx = 0; lx < len; lx++) buffer[lx] = (unsigned char)lrand (256);
	}
}


/* ------------------------------------------------------------
 * Function:    fillConstant
 * Description: Fill buffer with constant data
 * ------------------------------------------------------------ */
void fillConstant (unsigned char *buffer, unsigned char cData, long len)
{
if (blockPool > 0) {
	long l = lrand (blockPool);
	char *pPtr = pPool[l];
	memcpy (buffer, pPtr, len);
	}
else {
	for (long lx = 0; lx < len; lx++) buffer[lx] = cData;
	}
}


/* ------------------------------------------------------------
 * Function:    lsrand
 * Description: Seed local PRNGtwo buffers
 * ------------------------------------------------------------ */
static uint32 lseedx, lseedy;

void lsrand (unsigned int _seed)
{
lseedx = 2282008 + _seed;
lseedy = 362436069 - _seed;
}
        
        
/* ------------------------------------------------------------
 * Function:    lrand
 * Description: Local PRNG
 * ------------------------------------------------------------ */
long lrand (long maxRet)
{
lseedx = 69069 * lseedx + 123;
lseedy ^= lseedy << 13;
lseedy ^= lseedy >> 17;
lseedy ^= lseedy << 5;
return ((lseedx + lseedy) % maxRet);
}


/* ------------------------------------------------------------
 * Function:    getTime
 * Description: Get high res time from microboottime
 * ------------------------------------------------------------ */
double getTime (void)
{
double seconds;
 
#ifdef WIN32
LARGE_INTEGER ticksPerSecond;
LARGE_INTEGER tick;
QueryPerformanceFrequency (&ticksPerSecond);
QueryPerformanceCounter (&tick);
seconds = (double)tick.QuadPart / (double)ticksPerSecond.QuadPart;
#else
timespec ts;
clock_gettime (CLOCK_REALTIME, &ts);
seconds = ts.tv_sec + ((double)ts.tv_nsec / 1000000000.0);
#endif
return seconds;
}


/* ------------------------------------------------------------
 * Function:    lpretty
 * Description: Return number with commas
 * ------------------------------------------------------------ */
char *lpretty (long long value)
{
sprintf (pBuffer, "%llu", value);
return pretty (pBuffer);
}


/* ------------------------------------------------------------
 * Function:    llpretty
 * Description: Return number with commas
 * ------------------------------------------------------------ */
char *llpretty (long long value)
{
sprintf (pBuffer, "%llu", value);
return pretty (pBuffer);
}


/* ------------------------------------------------------------
 * Function:    pretty
 * Description: Return number with commas
 * ------------------------------------------------------------ */
char *pretty (char *buffer)
{        
char *ptr = strchr (buffer, 0) - 3, tmp[256];
while (ptr > buffer) {
        strcpy (tmp, ptr);
        *ptr = ',';
        strcpy (ptr+1, tmp);
        ptr -= 3;
        }
return buffer;
}


/* ------------------------------------------------------------
 * Function:	reduceValue
 * Description:	Reduce large value to XxxMB/s format
 * ------------------------------------------------------------ */
double reduceValue (char *scale, double r)
{
double rate = r;
strcpy (scale, "B/s");

if (rate >= 1024.0) {
	strcpy (scale, "KB/s"); rate = rate / 1024.0;
	if (rate >= 1024.0) {
		strcpy (scale, "MB/s"); rate = rate / 1024.0;
		if (rate >= 1024.0) {
			strcpy (scale, "GB/s"); rate = rate / 1024.0;
			}
		}
	}
return rate;
}


/* ------------------------------------------------------------
 * Function:	getNumber
 * Description:	Get a number with an optional scale (KB, MB, GB,
 *		TB). Long return.
 * ------------------------------------------------------------ */
long getNumber (char *buf)
{
long ret;
char *ptr;

ret = strtol (buf, &ptr, 10);
if (strncasecmp (ptr, "KB", 2) == 0) ret = ret * 1024;
else if (strncasecmp (ptr, "MB", 2) == 0) ret = ret * 1048576;
else if (strncasecmp (ptr, "GB", 2) == 0) ret = ret * 1073741824;
else if (strncasecmp (ptr, "TB", 2) == 0) ret = ret * 1024 * 1024 * 1024 * 1024;
return ret;
}


/* ------------------------------------------------------------
 * Function:	getNumber
 * Description:	Get a number with an optional scale (KB, MB, GB,
 *		TB). Long long return.
 * ------------------------------------------------------------ */
long long getNumberl (char *buf)
{
long long ret;
char *ptr;

ret = strtoll (buf, &ptr, 10);
if (strncasecmp (ptr, "KB", 2) == 0) ret = ret * 1024;
else if (strncasecmp (ptr, "MB", 2) == 0) ret = ret * 1048576;
else if (strncasecmp (ptr, "GB", 2) == 0) ret = ret * 1073741824;
else if (strncasecmp (ptr, "TB", 2) == 0) ret = ret * 1024 * 1024 * 1024 * 1024;
return ret;
}

/* ------------------------------------------------------------
 * Function:	getLongArg
 * Description:	Get a long input from the command line
 *		long long result
 * ------------------------------------------------------------ */
int getLongArg (char **argv, int x, long long *result)
{
int ret = x;
if(argv[ret][2] == 0) *result = getNumberl (argv[++ret]);
else *result= getNumberl (&argv[ret][2]);
return ret;
}


/* ------------------------------------------------------------
 * Function:	getLongArg
 * Description:	Get a long input from the command line
 *		long result
 * ------------------------------------------------------------ */
int getLongArg (char **argv, int x, long *result)
{
int ret = x;
if (argv[ret][2] == 0) *result = getNumber (argv[++ret]);
else *result = getNumber (&argv[ret][2]);
return ret;
}


/* ------------------------------------------------------------
 * Function:	strtoll
 * Description:	Convert string to long long, same calling format
 *				as strtol.
 * ------------------------------------------------------------ */
long long strtoll (char *buf, char **eos, int base)
{
long long ret = 0;
char *ptr = buf;

while (1) {
	if (*ptr >= '0' && *ptr <= '9') {
		ret *= 10;
		ret += *ptr - '0';
		}
	else break;
	ptr++;
	}
*eos = ptr;
return ret;
}


/* ------------------------------------------------------------
 * Function:	getFilesize
 * Description:	Return file's size
 * ------------------------------------------------------------ */
long long getFilesize (char *filename)
{
struct stat sb;

stat (filename, &sb);

return (long long)sb.st_size;
}


/* ------------------------------------------------------------
 * Function:	compareBuffers
 * Description:	Read loop helper - compare predicted buffer
 *		against actual buffer.
 * ------------------------------------------------------------ */
long compareBuffers (char *buffer, char *buffer1, long length)
{
long miscompares = 0, lx;
char *src, *tgt;

src = buffer; tgt = buffer1;

for (lx = 0; lx < length; lx++) {
	if (*src++ != *tgt++) miscompares++;
	}
return miscompares;
}


/* ------------------------------------------------------------
 * Function:	getDirCount
 * Description:	Count the number of subdirs matching
 *				subdir###### pattern.
 * ------------------------------------------------------------ */
long getDirCount (const char *base)
{
long dirCount = 0, ld;
char *filename = new char[8192];

#ifdef WIN32
WIN32_FIND_DATA FindFileData;
HANDLE hFind;
wchar_t *wFilename = new wchar_t[8192];
swprintf (wFilename, L"%s\\subdir*", base);
if ((hFind = FindFirstFile (wFilename, &FindFileData)) != INVALID_HANDLE_VALUE) {
	wcstombs (filename, FindFileData.cFileName, 8192);
	if ((ld = getNumber (filename + 6)) > dirCount) dirCount = ld;
	while (FindNextFile (hFind, &FindFileData) != 0) {
		wcstombs (filename, FindFileData.cFileName, 8192);
		if ((ld = getNumber (filename + 6)) > dirCount) dirCount = ld;
		}		
	FindClose (hFind);
	}
delete[] wFilename;
#else
DIR *dirp;
struct dirent *dp;
sprintf (filename, "%s", base);
if ((dirp = opendir (filename)) != (DIR *)NULL) {
	while ((dp = readdir (dirp)) != NULL) {
		if (strncasecmp (dp->d_name, "subdir", 6) == 0) {
			if ((ld = getNumber (dp->d_name + 6)) > dirCount) dirCount = ld;
			}
		}
	closedir (dirp);
	}
#endif
delete[] filename;
return dirCount + 1;			// Loops are <, so set this one higher
}


/* ------------------------------------------------------------
 * Function:	getFileCount
 * Description:	Count the number of subdirs matching
 *				subdir###### pattern.
 * ------------------------------------------------------------ */
long getFileCount (const char *base)
{
long fileCount = 0, ld;
char *filename = new char[8192];

#ifdef WIN32
WIN32_FIND_DATA FindFileData;
HANDLE hFind;
wchar_t *wFilename = new wchar_t[8192];
swprintf (wFilename, L"%s\\subdir000000\\file*.dat", base);
if ((hFind = FindFirstFile (wFilename, &FindFileData)) != INVALID_HANDLE_VALUE) {
	wcstombs (filename, FindFileData.cFileName, 8192);
	if ((ld = getNumber (filename + 4)) > fileCount) fileCount = ld;
	while (FindNextFile (hFind, &FindFileData) != 0) {
		wcstombs (filename, FindFileData.cFileName, 8192);
		if ((ld = getNumber (filename + 4)) > fileCount) fileCount = ld;
		}		
	FindClose (hFind);
	}
delete[] wFilename;
#else
DIR *dirp;
struct dirent *dp;
sprintf (filename, "%s\\subdir000000", base);
if ((dirp = opendir (filename)) != (DIR *)NULL) {
	while ((dp = readdir (dirp)) != NULL) {
		if (strncasecmp (dp->d_name, "file", 4) == 0) {
			ld = getNumber (dp->d_name + 4);
			if (ld > fileCount) fileCount = ld;
			}
		}
	closedir (dirp);
	}
#endif
delete[] filename;
return fileCount + 1;			// Loops are <, so set this one higher
}


/* ------------------------------------------------------------
 * Function:	openFile
 * Description:	Abstration to handle handles/streams opens
 * ------------------------------------------------------------ */
bool openFile (long openType, char *filename)
{
bool ret = true;

_fileOffset = 0;
_writeErrors = 0;
_readErrors = 0;
if (openType & GENERIC_READ) {
	if (useHandles) {
#ifdef WIN32
		wchar_t *wFilename = new wchar_t[4096];
		DWORD attributes = FILE_ATTRIBUTE_NORMAL;
		if (writeThru) attributes = attributes | /* FILE_FLAG_NO_BUFFERING | */ FILE_FLAG_WRITE_THROUGH;
		// NOTE: FILE_FLAG_NO_BUFFERING requires block I/O alignment - need to adjust blocksizes/filesizes so last write it aligned
		mbstowcs (wFilename, filename, 4096);
		if ((hFile = CreateFile (wFilename, GENERIC_READ, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE,
		  NULL, OPEN_EXISTING, attributes, NULL)) == INVALID_HANDLE_VALUE) {
			fwprintf (stderr, L"\nError: Cannot open %s for reading: ", wFilename); dispError (stderr);
			if (sout != (FILE *)NULL) {
				fwprintf (sout, L"%s, *** Cannot open for reading: ", wFilename); dispError (sout);
				}
			hFile = NULL;
			ret = false;
			}
		else {
			if (doubleOpen) {
				if ((hFile1 = CreateFile (wFilename, GENERIC_READ, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE,
				  NULL, OPEN_EXISTING, attributes, NULL)) == INVALID_HANDLE_VALUE) {
					fwprintf (stderr, L"\nError: Cannot open %s for reading (2nd): ", wFilename); dispError (stderr);
					if (sout != (FILE *)NULL) {
						fwprintf (sout, L"%s, *** Cannot open for reading (2nd): ", wFilename); dispError (sout);
						}
					hFile1 = NULL;
					}
				}
			}
		delete[] wFilename;
#endif
		}
	else {
		if ((fFile = fopen (filename, "rb")) == (FILE *)NULL) {
			fprintf (stderr, "Error: Cannot open %s for reading. (%s [%d])\n", filename, strerror (errno), errno);
			if (sout != (FILE *)NULL) fprintf (sout, "%s, *** Cannot open for reading: %s (%d)\n", filename, strerror (errno), errno);
			ret = false;
			}
		}
	}
else if (openType & GENERIC_WRITE) {
	if (useHandles) {
#ifdef WIN32
		DWORD openMode;
		DWORD attributes = FILE_ATTRIBUTE_NORMAL;
		if (appendMode) openMode = OPEN_ALWAYS; else openMode = CREATE_ALWAYS;
		if (writeThru) attributes = attributes | /* FILE_FLAG_NO_BUFFERING | */ FILE_FLAG_WRITE_THROUGH;
		wchar_t *wFilename = new wchar_t[4096];
		mbstowcs (wFilename, filename, 4096);
		if ((hFile = CreateFile (wFilename, GENERIC_WRITE, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE,
			NULL, CREATE_ALWAYS, attributes, NULL)) == INVALID_HANDLE_VALUE) {
			fwprintf (stderr, L"\nError: Cannot open %s for writing: ", wFilename); dispError (stderr);
			if (sout != (FILE *)NULL) {
				fwprintf (sout, L"%s, *** Cannot open for writing: ", wFilename);
				dispError (sout);
				}
			ret = false;
			hFile = NULL;
			}
		else {
			if (doubleOpen) {
				if ((hFile1 = CreateFile (wFilename, GENERIC_WRITE, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE,
				  NULL, CREATE_ALWAYS, attributes, NULL)) == INVALID_HANDLE_VALUE) {
					fwprintf (stderr, L"\nError: Cannot open %s for writing (2nd): ", wFilename); dispError (stderr);
					if (sout != (FILE *)NULL) {
						fwprintf (sout, L"%s, *** Cannot open for writing (2nd): ", wFilename); dispError (stderr);
						}
					hFile1 = NULL;
					}
				}
			}
		if (appendMode) {
			SetFilePointer (hFile, 0, NULL, FILE_END);		// Move to the end of file
			_fileOffset = (long long)GetFileSize (hFile, NULL);
			}
		delete[] wFilename;
#endif
		}
	else {
		if (appendMode) {
			if (access (filename, 0) == 0) {
				if ((fFile = fopen (filename, "rb+")) == (FILE *)NULL) {
					fprintf (stderr, "Error: Cannot open %s for apending. (%s [%d])\n", filename, strerror (errno), errno);
					if (sout != (FILE *)NULL) fprintf (sout, "%s, *** Cannot open for appending: %s (%d)\n", filename, strerror (errno), errno);
					ret = false;
					}
				else {
					fseek (fFile, 0L, SEEK_END);			// Move to EOF for writes
					_fileOffset = (long long)ftell (fFile);
					}
				}
			else {
				if ((fFile = fopen (filename, "wb")) == (FILE *)NULL) {
					fprintf (stderr, "Error: Cannot open %s for writing. (%s [%d])\n", filename, strerror (errno), errno);
					if (sout != (FILE *)NULL) fprintf (sout, "%s, *** Cannot open for writing: %s (%d)\n", filename, strerror (errno), errno);
					ret = false;
					}
				}
			}
		else {
			if ((fFile = fopen (filename, "wb")) == (FILE *)NULL) {
				fprintf (stderr, "Error: Cannot open %s for writing. (%s [%d])\n", filename, strerror (errno), errno);
				if (sout != (FILE *)NULL) fprintf (sout, "%s, *** Cannot open for writing: %s (%d)\n", filename, strerror (errno), errno);
				ret = false;
				}
			}
		}
	}
return ret;
}


/* ------------------------------------------------------------
 * Function:	closeFile
 * Description:	Abstration to handle closes
 * ------------------------------------------------------------ */
void closeFile (void)
{
if (useHandles) {
#ifdef WIN32
	if (hFile != NULL) {
		if (CloseHandle (hFile) == 0) {
			fwprintf (stderr, L"\nError: CloseHandle failed: "); dispError (stderr);
			if (sout != (FILE *)NULL) {
				fwprintf (sout, L"*** CloseHandle failed: ");
				dispError (sout);
				}
			}
		hFile = NULL;
		}
	if (hFile1 != NULL) {
		if (CloseHandle (hFile1) == 0) {
			fwprintf (stderr, L"\nError: CloseHandle of 2nd handle failed: "); dispError (stderr);
			if (sout != (FILE *)NULL) {
				fwprintf (sout, L"\n*** CloseHandle failed: ");
				dispError (sout);
				}
			}
		hFile1 = NULL;
		}
#endif
	}
else {
	if (fFile != NULL) {
		fclose (fFile);
		fFile = NULL;
		}
	}
}


/* ------------------------------------------------------------
 * Function:	writeFile
 * Description:	Abstration to handle writes
 * NOTE: Add inter-block sleep and/or reopen
 * ------------------------------------------------------------ */
long writeFile (unsigned char *buffer, long count)
{
long bytesWritten = 0;

errno = 0;

if (useHandles) {
#ifdef WIN32
	DWORD bytesRet;
	if (hFile != NULL) {
		if (WriteFile (hFile, buffer, (DWORD)count, &bytesRet, NULL) == 0) {
			_writeErrors++;
			if (sout != (FILE *)NULL) {
				fprintf (sout, "*** Write failure - offset:%lld request:%ld wrote:%ld: ", _fileOffset, count, bytesRet);
				dispError (sout);
				}
			}
		bytesWritten = (long)bytesRet;
		_fileOffset += bytesWritten;
		if (flushIO) FlushFileBuffers (hFile);
		}
#endif
	}
else {
	if (fFile != NULL) {
		bytesWritten = fwrite (buffer, 1, count, fFile);
		if (errno != 0) {
			_writeErrors++;
			if (sout != (FILE *)NULL) fprintf (sout, "*** Write failure - offset:%lld request:%ld wrote:%ld: %s (%d)\n", _fileOffset, count, bytesWritten, strerror (errno), errno);
			}
		_fileOffset += bytesWritten;
		if (flushIO) fflush (fFile);
		}
	}
return bytesWritten;
}


/* ------------------------------------------------------------
 * Function:	readFile
 * Description:	Abstration to handle reads
 * ------------------------------------------------------------ */
long readFile (unsigned char *buffer, long bytes)
{
long bytesRead = 0;

errno = 0;

if (useHandles) {
#ifdef WIN32
	DWORD count;
	if (hFile != NULL) {
		if (ReadFile (hFile, buffer, (DWORD)bytes, &count, NULL) == 0) {
			_readErrors++;
			if (sout != (FILE *)NULL) {
				fprintf (sout, "*** Read failure - offset:%lld requested:%ld read:%ld: ", _fileOffset, bytes, count);
				dispError (sout);
				}
			}
		bytesRead = (long)count;
		_fileOffset += bytesRead;
		}
#endif
	}
else {
	if (fFile != NULL) {
		bytesRead = fread (buffer, 1, bytes, fFile);
		if (errno != 0) {
			if (sout != (FILE *)NULL) fprintf (sout, "*** Read failure - offset:%lld requested:%ld read:%ld: %s (%d)\n", _fileOffset, bytes, bytesRead, strerror (errno), errno);
			}
		_fileOffset += bytesRead;
		}
	}
return bytesRead;
}


/* ------------------------------------------------------------
 * Function:	backup
 * Description:	Backspace x characters
 * ------------------------------------------------------------ */
void backup (int count)
{
for (int x = 0; x < count; x++) fputc (0x08, stdout);
}


/* ------------------------------------------------------------
 * Function:	compareDoubles
 * Description:	Qsort helper
 * ------------------------------------------------------------ */
int compareDoubles (const void *a, const void *b)
{
const double da = *(const double *)a;
const double db = *(const double *)b;
return (da > db) - (da < db);
}


/* ------------------------------------------------------------
 * Function:	usleep
 * Description:	ms sleep
 * ------------------------------------------------------------ */
void usleep (double sec)
{
long usec = (long)(sec * 1000000);
struct timeval tv;
tv.tv_sec = usec / 1000000;
tv.tv_usec = usec % 1000000;
select (0, NULL, NULL, NULL, &tv);
}


/* ------------------------------------------------------------
 * Function:	dispError
 * Description:	Let Windows gen the error from GetLastError
 * ------------------------------------------------------------ */
#ifdef WIN32
void dispError (FILE *fout)
{
char *buffer = new char[1024], *ptr;
wchar_t *wBuffer = new wchar_t[1024];
DWORD err = GetLastError ();

FormatMessage (FORMAT_MESSAGE_FROM_SYSTEM, NULL, err, 0, wBuffer, 1024, NULL);
wcstombs (buffer, wBuffer, 1024);
if ((ptr = strchr (buffer, 0x0a)) != (char *)NULL) *ptr = 0;
if ((ptr = strchr (buffer, 0x0d)) != (char *)NULL) *ptr = 0;
fprintf (fout, "%s (%d)\n", buffer, err);
delete[] buffer;
delete[] wBuffer;
}
#endif
