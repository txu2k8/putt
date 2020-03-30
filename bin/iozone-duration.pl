#!/usr/bin/perl -w

##############################################################################
# Copyright (C) 2008-2011 VMware, Inc.
# All Rights Reserved
##############################################################################
# Author: Chris Waago
# Last Modified By: hkushnireuskaya
# Date Last Modified: 6/23/2011
##############################################################################

use strict;
use Getopt::Long;

my $duration = 3600;            # test duration in seconds
my $testfile = "/iozone.dat";   # iozone -f option (test file path)
my $fileSize = "100m";          #        -q option (max file size)
my $recSize = "4096k";          #        -g option (max record size)
my $help = undef;               # option to print help

my $cmd = undef;                # iozone command to execute

sub CheckInput;                 # check user input
sub PrintTestParams;            # print test parameters
sub Run;                        # run the test
sub Help;                       # print help and exit


GetOptions (
            'duration=i'   => \$duration,
            'testfile=s'   => \$testfile,
            'filesize=s'   => \$fileSize,
            'recordsize=s' => \$recSize,
            'help'         => \$help
           );

if ( $help ) {
    Help;
}

CheckInput;
PrintTestParams;
Run;


###################
# Check user input
###################

sub CheckInput
{
    my $gos = undef;       # guest type (win or lin)
    my $errors = 0;        # count errors

    unless ( $duration and $testfile and $fileSize and $recSize ) {
        print "Error: Must specify all input parameters, exiting. \n";
        exit;
    }

    if ( $^O eq 'linux' ) {
        $gos = "lin";   # also esx
    } else {
        $gos = "win";
    }

    if ( $gos eq 'lin' ) {
        $cmd =
          "/usr/bin/iozone -a -+d -f $testfile -q $recSize -g $fileSize -o -D -Rb ./iozone.xls";
        unless ( -e "/usr/bin/iozone" ) {
            print "Warning: IOzone executable (/opt/iozone/bin/iozone) does " .
                  "not exist. \n";
            unless ( -e "/iozone/bin/iozone" ) {
                print "Error: IOzone executable (/iozone/bin/iozone) does " .
                    "not exist. \n";
                $errors++;
            } else {
                $cmd =
                "/iozone/bin/iozone -a -+d -f $testfile -q $recSize -g $fileSize";
            }

        }
    } else {
        $cmd =
          "/iozone/iozone.exe -a -+d -f $testfile -q $recSize -g $fileSize";
        unless ( -e "/iozone/iozone.exe" ) {
            print "Error: IOzone executable (c:/iozone/iozone.exe) does not " .
                  "exist. \n";
            $errors++;
        }
    }

    if ( $testfile =~ /(.*(\/|\\)).*/ ) {
        my $testDir = $1;
        unless ( -e "$testDir" ) {
            print "Error: Test folder ($testDir) does not exist. \n";
            $errors++;
        }
    }

    if ( $errors > 0 ) {
        print "Errors detected, exiting. \n";
        exit;
    }
}


########################
# Print test parameters
########################

sub PrintTestParams
{
    my $hostname = `hostname`;
    chomp ( $hostname );

    my ($second, $minute, $hour, $day, $month, $year,
        $wday, $yday, $isdst) = localtime();
    printf "Test start time: %d-%02d-%02d %02d:%02d:%02d \n", $year + 1900,
        $month + 1, $day, $hour, $minute, $second;

    print "Test Parameters: \n" .
          "Server: $hostname \n" .
          "Command: $cmd \n" .
          "Duration: $duration seconds \n" .
          "Test File: $testfile \n" .
          "File Size: $fileSize \n" .
          "Record Size: $recSize \n";
}

#############
# Run iozone
#############

sub Run 
{
    my $start = time;         # mark start of duration
    my $startTime = 0;        # start time before individual test
    my $elapsedTime = 0;      # elapsed time of individual test
    my $elapsed = 0;          # total elapsed time
    my $count = 0;            # count the iterations

    while ( $elapsed < $duration ) {
        $count++;
        $startTime = time;
        system "$cmd";
        sleep 2;
        $elapsedTime = time - $startTime;

        $elapsed = time - $start;
        print "\nCompleted iteration $count in $elapsedTime seconds. \n\n";
        if ( ( -e "C:\\terminate_io" ) or ( -e "/tmp/terminate_io") ) {
           print "Gracefull exit \n\n";
           last;
        }

    }

    print "Total iterations: $count \n";
    print "Total elapsed time: $elapsed seconds \n";

    return 1;
}


#############
# Print help
#############

sub Help 
{
    print "Options: \n" .
      "  -d|--duration     duration in seconds (default=3600) \n" .
      "  -t|--testfile     path to test file (default=/tmp/iozone) \n" .
      "  -f|--filesize     max file size (default=1g) \n" .
      "  -r|--recordsize   max record size (default=4096k) \n" .
      "  -h|--help         print this help \n";

    print "\nExamples: " .
      " ./iozone-duration.pl \n" .
      " ./iozone-duration.pl -d 600 -t /data/iozone -f 150m -r 2048k \n" .
      " ./iozone-duration.pl -d 7200 -t e:\\data\\iozone -f 500m -r 1024k \n\n";

    exit;
}

