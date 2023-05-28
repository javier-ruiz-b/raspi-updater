# Raspi updater

![build-and-test](https://github.com/javier-ruiz-b/raspi-updater/actions/workflows/build-and-test.yml/badge.svg)

A Raspberry Pi image updater which uses quic-go for secure transfers over the network.

The updater is intended to be run from the initramfs.

WARNING: This is a WIP and specific to a use case. Use at your own risk.


## Requirements

    * initramfs-tools Debian package
    * Same compression tool in server and client.
        * i.e: lz4, gzip, zstd, xz
    * Server with a hostname (for secure QUIC connetion)    

## How it works

### Setup

The Debian package installs an initramfs-hook. initramfs-tools is expected to
 be installed and running in the system.

The command **raspi-updater-config** configures the Raspberry client connection to
 the server. This command should be run to configure the client the first time.
 The raspi-updater server should run in any architecture supported by Go.


### Action

The client first backs up the contents of the SD card to the server before 
 writting the image. 

The rootfs and optional additional partitions are written first and the boot
 partition is written last to ensure the update to be restarted if something
 goes wrong.


### Example client output

```
=== RUN   TestAcceptance
2023/05/27 17:09:17 Checking for client update
2023/05/27 17:09:17 Reading local disk
2023/05/27 17:09:17 Reading local version
2023/05/27 17:09:17 Warning: could not read local version: boot partition not found
2023/05/27 17:09:17 Matching compression tools
2023/05/27 17:09:18 Using lz4fast compression
2023/05/27 17:09:18 Backing up disk if necessary
⠸ Saving backup acceptance (24 kB, 71 kB/s) [0s]  | (6.9/64 MB, 23 MB/s) [0s:2s]
Backup /tmp/acceptance3922856950/client.img  26% || (17/64 MB, 24 MB/s) [0s:1s] 
Backup /tmp/acceptance3922856950/client.img  81% || (52/64 MB, 31 MB/s) [1s:0s] 
Backup /tmp/acceptance3922856950/client.img 100% || (64/64 MB, 32 MB/s)         
⠋ Saving backup acceptance (257 kB, 128 kB/s) [2s] 
2023/05/27 17:09:20 Checking for image update for acceptance                    
2023/05/27 17:09:20 Image update available.
  Server version: 1.0
  Client version:
2023/05/27 17:09:20 Getting partition scheme
2023/05/27 17:09:20 Downloading boot partition
Sending acceptance partition 0  31% |████         | (11/36 MB, 33 M            
⠴ Download boot partition (68 kB, 115 kB/s) [0s]  | (18/36 MB, 28 MB/s) [0s:0s]
Sending acceptance partition 0  69% |████████     | (25/36 MB, 24 MB/s) [1s:0s]    
Sending acceptance partition 0 100% |█████████████| (36/36 MB, 18 MB/s)        
⠏ Download boot partition (145 kB, 74 kB/s) [1s]                               
2023/05/27 17:09:22 Backing up disk if necessary
2023/05/27 17:09:22 Backup exists on server.
2023/05/27 17:09:22 Merging partition table
2023/05/27 17:09:22 Partition table:
Local: Partition table TotalSize 67 MB, SectorSize 512

Remote: Partition table TotalSize 64 MB, SectorSize 512
 1; PartType 0x0b,  TotalSize   38 MB,  StartSec     2048,  EndSec    75776
 2; PartType 0x83,  TotalSize   25 MB,  StartSec    75776,  EndSec   124928

Final: Partition table TotalSize 64 MB, SectorSize 512
 1; PartType 0x0b,  TotalSize   38 MB,  StartSec     2048,  EndSec    75776
 2; PartType 0x83,  TotalSize   25 MB,  StartSec    75776,  EndSec   124928
2023/05/27 17:09:22 Writing partition table
2023/05/27 17:09:22 Running partprobe
2023/05/27 17:09:22 Faking successful /bin/partprobe  execution
2023/05/27 17:09:22 Reading disk
2023/05/27 17:09:22 Writing partitions
Sending acceptance partition 1  14% |█          |                 
Sending acceptance partition 1  41% |█                                         
Sending acceptance partition 1  65% |███████                                   
Sending acceptance partition 1  79% |███                                       
Transferring partition 2 / 2  63% |█████████      | (15/24 MB, 12 MB/s) [1s:0s]
Sending acceptance partition 1 100% |█████████████| (24/24 MB, 11 MB/s)        
Transferring partition 2 / 2 100% |███████████████| (24/24 MB, 14 MB/s)        
2023/05/27 17:09:25 Faking successful /bin/sync  execution
2023/05/27 17:09:25 Writing boot partition                                     
Writing boot partition  26% |██                                                   
Writing boot partition  61% |████████████         | (22/36 MB, 17 MB/s) [1s:0s]
Writing boot partition  80%                                                    
Writing boot partition 100% |█████████████████████| (36/36 MB, 20 MB/s)        
2023/05/27 17:09:27 Faking successful /bin/sync  execution
2023/05/27 17:09:27 Couldn't write version 1.0 after update                    
2023/05/27 17:09:27 Faking successful /bin/sync  execution
2023/05/27 17:09:27 Update complete. Rebooting in 5 seconds
2023/05/27 17:09:27 Faking successful /usr/bin/sleep 5 execution
2023/05/27 17:09:27 Faking successful /usr/bin/busybox reboot -f execution
2023/05/27 17:09:27 Images: /tmp/acceptance3922856950/acceptance_1.0.img /tmp/acceptance3922856950/client.img
--- PASS: TestAcceptance (16.04s)
PASS
```