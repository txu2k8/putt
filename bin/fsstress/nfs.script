#!/bin/sh

DURATION=$1

HOSTNAME=$(hostname | awk {'print $1'})

#Convert to seconds
DURATION=$(( $DURATION * 60 * 60 ))

mkdir -p /mnt/nfs/ext3/udp/2/${HOSTNAME}1
mkdir -p /mnt/nfs/ext3/udp/3/${HOSTNAME}2
mkdir -p /mnt/nfs/ext3/tcp/2/${HOSTNAME}3
mkdir -p /mnt/nfs/ext3/tcp/3/${HOSTNAME}4
mkdir -p /mnt/nfs/jfs/udp/2/${HOSTNAME}1
mkdir -p /mnt/nfs/jfs/udp/3/${HOSTNAME}2
mkdir -p /mnt/nfs/jfs/tcp/2/${HOSTNAME}3
mkdir -p /mnt/nfs/jfs/tcp/3/${HOSTNAME}4
mkdir -p /mnt/nfs/reiserfs/udp/2/${HOSTNAME}1
mkdir -p /mnt/nfs/reiserfs/udp/3/${HOSTNAME}2
mkdir -p /mnt/nfs/reiserfs/tcp/2/${HOSTNAME}3
mkdir -p /mnt/nfs/reiserfs/tcp/3/${HOSTNAME}4

./fsstress -l 0 -d /mnt/nfs/ext3/udp/2/${HOSTNAME}1 -n 1000 -p 50 -r > ext3.udp.2.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/ext3/udp/3/${HOSTNAME}2 -n 1000 -p 50 -r > ext3.udp.3.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/ext3/tcp/2/${HOSTNAME}3 -n 1000 -p 50 -r > ext3.tcp.2.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/ext3/tcp/3/${HOSTNAME}4 -n 1000 -p 50 -r > ext3.tcp.3.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/jfs/udp/2/${HOSTNAME}1 -n 1000 -p 50 -r > jfs.udp.2.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/jfs/udp/3/${HOSTNAME}2 -n 1000 -p 50 -r > jfs.udp.3.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/jfs/tcp/2/${HOSTNAME}3 -n 1000 -p 50 -r > jfs.tcp.2.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/jfs/tcp/3/${HOSTNAME}4 -n 1000 -p 50 -r > jfs.tcp.3.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/reiserfs/udp/2/${HOSTNAME}1 -n 1000 -p 50 -r > reiserfs.udp.2.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/reiserfs/udp/3/${HOSTNAME}2 -n 1000 -p 50 -r > reiserfs.udp.3.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/reiserfs/tcp/2/${HOSTNAME}3 -n 1000 -p 50 -r > reiserfs.tcp.2.log 2>&1 &
./fsstress -l 0 -d /mnt/nfs/reiserfs/tcp/3/${HOSTNAME}4 -n 1000 -p 50 -r > reiserfs.tcp.3.log 2>&1 &

sar -o ./nfs.sardata 30 0 &

echo "Test set for $DURATION seconds"
echo "Testing in progress"
sleep $DURATION
echo "Testing done. Killing processes"
killall -9 sadc
killall -9 fsstress
ps -ef | grep -v grep | grep fsstress > /dev/null 2>&1
TESTING=$?
while [ $TESTING -eq 0 ]
do
  killall -9 fsstress
  echo -n "."
  sleep 5
  ps -ef | grep -v grep | grep fsstress > /dev/null 2>&1
  TESTING=$?
done 
echo "All processes killed. Done."


