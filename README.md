Here is the English translation of the README.md document:

***

# BaiduPCS-Go Baidu Netdisk Client (Enhanced Version)


A command-line client for Baidu Netdisk that mimics Linux shell file processing commands.

iikira/BaiduPCS-Go was largely inspired by [GangZhuo/BaiduPCS](https://github.com/GangZhuo/BaiduPCS) and this project was largely based on iikira/BaiduPCS-Go.

## Notice

This version continues development based on iikira's original BaiduPCS-Go v3.6.2 and adds a transfer (saving shared files) function.

This software does not provide download speed acceleration beyond the official client. For configuration suggestions for normal users and SVIP users, please refer to [Display and modify program configuration items](#display-and-modify-program-configuration-items).

<!-- toc -->
## Table of Contents

- [Features](#features)
- [Version Updates](#version-updates)
- [Compilation/Cross-Compilation Instructions](#compilationcross-compilation-instructions)
- [Download/Run Instructions](#downloadrun-instructions)
  * [Installation](#installation)
  * [Windows](#windows)
  * [Linux / macOS](#linux--macos)
  * [Android / iOS](#android--ios)
- [Command List and Instructions](#command-list-and-instructions)
  * [Attention !!!](#attention-)
  * [Check for program updates](#check-for-program-updates)
  * [Login to Baidu Account](#login-to-baidu-account)
  * [List accounts](#list-accounts)
  * [Get current account](#get-current-account)
  * [Switch Baidu Account](#switch-baidu-account)
  * [Logout of Baidu Account](#logout-of-baidu-account)
  * [Get Netdisk Quota](#get-netdisk-quota)
  * [Switch working directory](#switch-working-directory)
  * [Print working directory](#print-working-directory)
  * [List directory](#list-directory)
  * [List directory tree](#list-directory-tree)
  * [Get file/directory meta information](#get-filedirectory-meta-information)
  * [Search files](#search-files)
  * [Download files/directories](#download-filesdirectories)
  * [Upload files/directories](#upload-filesdirectories)
  * [Get download direct link](#get-download-direct-link)
  * [Repair file MD5](#repair-file-md5)
  * [Create directory](#create-directory)
  * [Delete files/directories](#delete-filesdirectories)
  * [Copy files/directories](#copy-filesdirectories)
  * [Move/Rename files/directories](#moverename-filesdirectories)
  * [Transfer files/directories](#transfer-filesdirectories)
  * [Share files/directories](#share-filesdirectories)
    + [Set share for files/directories](#set-share-for-filesdirectories)
    + [List shared files/directories](#list-shared-filesdirectories)
    + [Cancel share for files/directories](#cancel-share-for-filesdirectories)
  * [Offline Download](#offline-download)
    + [Add offline download task](#add-offline-download-task)
    + [Query offline download task precisely](#query-offline-download-task-precisely)
    + [Query offline download task list](#query-offline-download-task-list)
    + [Cancel offline download task](#cancel-offline-download-task)
    + [Delete offline download task](#delete-offline-download-task)
  * [Recycle Bin](#recycle-bin)
    + [List recycle bin files](#list-recycle-bin-files)
    + [Restore recycle bin files or directories](#restore-recycle-bin-files-or-directories)
    + [Delete recycle bin files or directories / Empty recycle bin](#delete-recycle-bin-files-or-directories--empty-recycle-bin)
  * [Display and modify program configuration items](#display-and-modify-program-configuration-items)
  * [Test wildcards](#test-wildcards)
  * [Toolbox](#toolbox)
- [Beginner Tutorial](#beginner-tutorial)
  * [1. View program usage instructions](#1-view-program-usage-instructions)
  * [2. Login to Baidu Account (Required)](#2-login-to-baidu-account-required)
  * [3. Switch Netdisk working directory](#3-switch-netdisk-working-directory)
  * [4. List files and directories in Netdisk](#4-list-files-and-directories-in-netdisk)
  * [5. Download files](#5-download-files)
  * [6. Set maximum download concurrency](#6-set-maximum-download-concurrency)
  * [7. Restore default configuration](#7-restore-default-configuration)
  * [8. Exit program](#8-exit-program)
- [Known Issues](#known-issues)
- [TODO](#todo)
- [Communication and Feedback](#communication-and-feedback)

<!-- tocstop -->

# Features

Multi-platform support, supports Windows, macOS, Linux, mobile devices, etc.

Baidu account multi-user support;

Wildcard matching for Netdisk paths and Tab auto-completion for commands and paths, [Wildcard - Baidu Baike](https://baike.baidu.com/item/通配符);

[Download](#download-filesdirectories) files inside Netdisk, supports downloading multiple files or directories, supports resuming broken downloads and parallel downloading of single files;

[Upload](#upload-filesdirectories) local files, supports uploading large files up to 128G, supports uploading multiple files or directories;

[Transfer](#transfer-filesdirectories) files shared by other users, supports shared links with passwords;

[Offline Download](#offline-download), supports http/https/ftp/eD2k/Magnet links protocols.

# Version Updates

**2025.10.29** v4.0.0
- Upload re-supports skipping rapid upload `--norapid`
- Upload same-name file overwrite policy `--policy` supports `skip`, `overwrite`, `rsync`; supports `config` to set global default policy
- Due to interface changes, upload no longer supports resuming broken uploads, download is unaffected
- Added `config` setting `proxy_hostnames`, overseas VPS users experiencing upload issues can try configuring a return-to-China proxy for `pan.baidu.com`
- Other detail optimizations

**2025.08.30** v3.9.9
- Maximum single file upload support increased to 128G
- Upload speed optimization
- Download file pre-allocation cancelled
- Due to official interface changes, file upload forces rapid upload calculation
- transfer command fixed `--download` parameter

**2025.08.29** v3.9.8
- Fully fixed file upload issues
- Fully fixed file download issues
- Removed some unusable features and unsupported parameters

**2025.01.07** v3.9.7
- fix #359, #360
- fix #339

**2024.12.14** v3.9.6
- Disabled rapid upload transfer function
- Fixed regular transfer failure

**2023.09.30** v3.9.5
- Restored rapid upload transfer function, requires setting accessToken before use, see setastoken --help
- Local file upload using rapid upload does not require accessToken
- fix #301
- fix #302

**2023.09.06** v3.9.5-beta
- Restored rapid upload transfer (supports long and short links), thanks to tousakarin, the Tampermonkey script developer, for the contribution
- The new rapid upload interface requires developer authorization, stability unknown. This test version is for users with a strong need for rapid upload to try, please update with caution

**2023.09.05** v3.9.4
- fix #244, fixed occasional crash during resumed upload
- Optimized processing logic when local rapid upload fails

**2023.08.26** v3.9.3
- Due to official interface blocking rapid upload at the principle level, the rapid upload transfer function is cancelled
- Updated part of the usage instructions
- It is recommended for users using the file upload function to update to this version

**2023.06.03** v3.9.2
- Fixed rapid upload link unable to transfer, due to official interface changes rapid upload no longer supports short link format
- Fixed upload file unable to use rapid upload
- fix #254 supports -f parameter to output shared link with password
- fix #251 Added md5 decryption function provided by mengzonefire

**2023.03.19** v3.9.1
- Fixed rapid upload transfer returning error code 9019

**2022.12.04** v3.9.0:
- Optimized transfer error prompts
- fix #239
- update go version to 1.18

**2022.11.25** v3.8.9:
- fix #234, continued fix for unable to transfer files

**2022.11.12** v3.8.8:
- fix #234, fixed unable to transfer files

**2022.2.18** v3.8.7:
- fix #175, perform file size detection before formal upload

**2022.2.14** v3.8.6:
- fix #160 #173, fixed bug where uploading resulted in empty files
- fix #165, supports transfer links with built-in extraction codes
- fix #175, upload added -policy=rsync strategy, use with --norapid to only skip files whose size has not changed
- In view of #172, it is suggested that the maximum number of download threads should not exceed 12

**2022.1.1** v3.8.5:
#### This version has known issues that will cause file upload failures and empty files, it is recommended to skip this update
- Happy New Year 2022, this update adds many features, welcome to test
- fix #146, advance the detection of files with the same name in fail and skip upload policies (has issues)
- fix #158, config can configure to turn off file name legality detection
- fix #141, download added --mtime option to keep file modification time
- fix #130, config can configure force_login_username to force login to a specified username
- Automatic switch when the first download link is unavailable, increasing download success rate

**2021.10.6** v3.8.4:
- fix possible memory overflow during login
- Upload file names allowed to contain single quotes

**2021.8.27** v3.8.3:
- fix replaced default panUA to solve SVIP speed limiting
- fix removed invalid rapid upload repair function
- Optimized rapid upload logic to improve success rate
- Optimized rapid upload export logic to improve export success rate for new files

**2021.7.20** v3.8.2:
- fix reading large amounts of file information prone to timeout
- fix parsing error when rapid upload link filename contains "#"
- share list added share download count display
- config added configuration: Upload same name file processing strategy

**2021.6.9** v3.8.1:
- fix some old links unable to transfer
- Added upload same name file auto-skip option

**2021.5.21** v3.8.0:
- fix upload automatically rolls back around 100M (pending test)
- fix individual normal rapid upload links unable to transfer
- fix file name containing percent sign causes export exception
- Optimized upload retry strategy (pending test)

**2021.4.14** v3.7.9:
- fix abnormal exit during upload causing inability to load resume info
- fix upload occasionally stuck at 0B/s
- Pre-check file name legality before uploading
- Online update uses mirror source acceleration

**2021.3.20** v3.7.8:

- Optimized upload output information format
- Optimized upload logic, improved upload speed
- transfer added --fix parameter, can transfer blocked rapid upload links (inspired by [dupan-rapid-extract](https://github.com/mengzonefire/dupan-rapid-extract))

**2021.3.11** v3.7.7:

- fix error caused by trailing `/` when moving and renaming files
- fix online upgrade invalid after v3.7.2
- fix transfer false report of missing STOKEN

**2021.2.23** v3.7.6:

- fix download file reporting `x509: certificate is valid` error
- Improved capture of download error types
- download added --fullpath parameter, local directory retains complete structure starting from root directory of Netdisk

**2021.2.8** v3.7.5:

- fix sometimes false report of missing stoken
- fix rapid upload link transfer failure on Windows platform
- fix sometimes pcs request missing Host
- When shared link contains multiple files/directories, optional archive to directory named after the first file (rapid upload not supported)

**2021.1.31** v3.7.4:

- fix downloading directory loses directory structure
- fix share list status information display error
- Support custom file upload server

**2021.1.22** v3.7.3:

- Share supports custom share code and valid days
- Transfer supports automatic download to default directory after transfer is complete
- Added restore default configuration function
- tree command supports specifying maximum output layers and output with fsid

**2021.1.9** v3.7.2:

- Basically fixed login verification failure issue ([#15](https://github.com/qjfoidnh/BaiduPCS-Go/issues/15))
- Optimized download module implementation strategy, further improving download speed while ensuring stability (need to modify as suggested in [Display and modify program configuration items](#display-and-modify-program-configuration-items))
- update function restored, online upgrades are possible now
- Supports exporting rapid upload links without writing to file, directly output to console; supports generic rapid upload format export, see export --help
- Other bug fixes

**2021.1.2** v3.7.1:

- Supported multi-file concurrent upload, file concurrency number and single file shard number can be specified in configuration
- Fixed issue where max simultaneous download file config did not take effect
- Corrected some display and help errors

**2020.12.19** v3.7.0:

* Replaced invalid repository of iikira version
* Transfer function supports old short links
* Download file verification turned off by default, can be enabled in config file
* Fixed false report of download failure when verification is turned off
* Transfer function now supports username/password login and bduss login in addition to cookies login; bduss login requires specifying stoken simultaneously

**2020.11.08** v3.6.3:

* Fixed transfer failure
* Fixed share file failure


# Compilation/Cross-Compilation Instructions
Set the GOOS and GOARCH environment variables,

Run `go tool dist list` to see all supported GOOS/GOARCH.

## Linux/Darwin Example: Compile 64-bit program for Windows
```
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build
```
## Windows Example: Compile 32-bit program for Linux
```
set GOOS=linux
set GOARCH=386
set CGO_ENABLED=0
go build
```

# Download/Run Instructions

Go language program, compiled programs for common platforms can be downloaded directly from [Lanzou Cloud](https://wws.lanzoui.com/b01berebe). Password: 4pix

If the program outputs garbled characters when running, please check if the terminal encoding is `UTF-8`.

Before using this program, it is recommended to learn some Linux basics and basic commands.

If the program is run without any parameters, it will enter a cli interactive mode imitating a Linux shell system user interface, where relevant commands can be run directly.

In cli interactive mode, the prefix of the line where the cursor is located should be `BaiduPCS-Go >`, if logged into a Baidu account the format is `BaiduPCS-Go:<working directory> <Baidu ID>$ `.

The program will provide usage instructions for relevant commands.

## Installation


## Windows

The program should be run in Command Prompt or PowerShell, there may be display issues in mintty (e.g.: GitBash).

You can also double-click the program to run it. For specific usage, please refer to [Command List and Instructions](#command-list-and-instructions) and [Beginner Tutorial](#beginner-tutorial).

## Linux / macOS

The program should be run in a Terminal.

For specific usage, please refer to [Command List and Instructions](#command-list-and-instructions) and [Beginner Tutorial](#beginner-tutorial).

## Android / iOS

> Android / iOS mobile device operation is relatively troublesome, using this program on mobile devices is not recommended. Mobile devices cannot directly use the pre-compiled Linux arm64 version, users need to download the source code and cross-compile it themselves.

For Android, it is recommended to use [Termux](https://termux.com) or [NeoTerm](https://github.com/NeoTerm/NeoTerm) or a terminal emulator to provide a terminal environment.

Example: [Android Running This Project Program Reference Example](https://web.archive.org/web/20190820154934/https://github.com/iikira/BaiduPCS-Go/wiki/Android-%E8%BF%90%E8%A1%8C%E6%9C%AC%E9%A1%B9%E7%9B%AE%E7%A8%8B%E5%BA%8F%E5%8F%82%E8%80%83%E7%A4%BA%E4%BE%8B), those interested can refer to it.

For Apple iOS, jailbreak is required, search and download MobileTerminal in Cydia, or other software that provides a terminal environment.

Example: [iOS Running This Project Program Reference Example](https://web.archive.org/web/20190820155025/https://github.com/iikira/BaiduPCS-Go/wiki/iOS-%E8%BF%90%E8%A1%8C%E6%9C%AC%E9%A1%B9%E7%9B%AE%E7%A8%8B%E5%BA%8F%E5%8F%82%E8%80%83%E7%A4%BA%E4%BE%8B), those interested can refer to it.

For specific usage, please refer to [Command List and Instructions](#command-list-and-instructions) and [Beginner Tutorial](#beginner-tutorial).

# Command List and Instructions

## Attention !!!

The command prefix `BaiduPCS-Go` is the full path name pointing to the program run (the first argument of ARGv).

When running the program directly without any other parameters, the program enters cli interactive mode. When running the following commands, remove the command prefix `BaiduPCS-Go`!

Cli interactive mode supports pressing the tab key to auto-complete commands and paths.

## Check for program updates
```
BaiduPCS-Go update
```

## Login to Baidu Account

### Standard Login to Baidu Account

Supports online verification of bound mobile number or email.
Note: This method has not been maintained for a long time, it is recommended to use other login methods.
```
BaiduPCS-Go login
```

### Login using Baidu BDUSS and Baidu Netdisk STOKEN

[About getting Baidu BDUSS](https://blog.csdn.net/ykiwmy/article/details/103730962) The way to get STOKEN is basically the same as BDUSS. Note that STOKEN must be obtained on the Baidu Netdisk page, otherwise it is invalid.
STOKEN is a field in cookies, note it is not bdstoken, if the STOKEN you get has no uppercase letters, you probably got it wrong.

```
BaiduPCS-Go login -bduss=<BDUSS> -stoken=<STOKEN>
```

### Login using Baidu Cookies (Recommended)

[About getting Baidu Cookies](https://jingyan.baidu.com/article/5553fa829a6a9e65a23934b0.html)
The tutorial shows getting Cookies for Baidu Experience, here just change it to the Baidu Netdisk homepage.

```
BaiduPCS-Go login -cookies=<Cookies>
```

#### Examples
```
BaiduPCS-Go login -bduss=1234567 -stoken=234567
```
```
BaiduPCS-Go login # Interactive login is no longer maintained, not recommended
Please enter Baidu username (mobile number/email/username), press Enter to submit > 1234567
```
```
BaiduPCS-Go login -cookies="BAIDUID=50949C0890YG9735EA6Q3870AFE38:FG=1; BIDUPSID=112335C0ACCAFFJW675EA69A870AFE38; PSTM=1981928511; BDORZ=D6745EBF6F3SW24E515D22A1598; PANWEB=1; BDUSS=ASAYUGFHSTFKGBGSU; STOKEN=gfsdge9gisfgspig34254d7879eee5756b10sgeyrw5vyw342td510ffc9414d32251; SCRC=cwrywec5evyetra26bvvehefvfg6a8; BDCLND=C%4sfgGysrZ%2BML6; PANPSC=wreyewygdfhdggedhsdfg4353"
```

## List accounts

```
BaiduPCS-Go loglist
```

Lists all logged-in Baidu accounts.

## Get current account

```
BaiduPCS-Go who
```

## Switch Baidu Account

Switch logged-in Baidu account.
```
BaiduPCS-Go su <uid>
```
```
BaiduPCS-Go su

Please enter the # value of the account to switch to >
```

## Logout of Baidu Account

Logout of currently logged-in Baidu account.
```
BaiduPCS-Go logout
```

The program will further confirm the logout to prevent accidental operation.

## Get Netdisk Quota

```
BaiduPCS-Go quota
```
Get the total storage space and used storage space of the Netdisk.

## Switch working directory
```
BaiduPCS-Go cd <directory>
```

### Automatically list files and directories under the working directory after switching
```
BaiduPCS-Go cd -l <directory>
```

#### Examples
```
# Switch to /My Resources working directory
BaiduPCS-Go cd /My Resources

# Switch to parent directory
BaiduPCS-Go cd ..

# Switch to root directory
BaiduPCS-Go cd /

# Switch to /My Resources working directory, and automatically list files and directories under /My Resources
BaiduPCS-Go cd -l My Resources

# Use wildcards
BaiduPCS-Go cd /My*
```

## Print working directory
```
BaiduPCS-Go pwd
```

## List directory

List files and directories in the current working directory or specified directory.
```
BaiduPCS-Go ls
```
```
BaiduPCS-Go ls <directory>
```

### Optional Parameters
```
-asc: Sort ascending
-desc: Sort descending
-time: Sort by time
-name: Sort by filename
-size: Sort by size
```

#### Examples
```
# List files and directories inside My Resources
BaiduPCS-Go ls My Resources

# Absolute path
BaiduPCS-Go ls /My Resources

# Sort descending
BaiduPCS-Go ls -desc My Resources

# Sort descending by file size
BaiduPCS-Go ls -size -desc My Resources

# Use wildcards
BaiduPCS-Go ls /My*
```

## List directory tree

List a tree diagram of files and directories in the current working directory or specified directory.
```
BaiduPCS-Go tree <directory>

# Default get working directory meta information
BaiduPCS-Go tree
```

## Get file/directory meta information
```
BaiduPCS-Go meta <file/directory 1> <file/directory 2> <file/directory 3> ...

# Default get working directory meta information
BaiduPCS-Go meta
```

#### Examples
```
BaiduPCS-Go meta My Resources
BaiduPCS-Go meta /
```

## Search files

Search for files by filename (searching for directories is not supported).

Searches in the current working directory by default.

```
BaiduPCS-Go search [-path=<directory to search>] [-r] <keyword>
```

#### Examples
```
# Search for files in root directory
BaiduPCS-Go search -path=/ keyword

# Search for files in current working directory
BaiduPCS-Go search keyword

# Recursively search for files in current working directory
BaiduPCS-Go search -r keyword
```

## Download files/directories
```
BaiduPCS-Go download <Netdisk file or directory path 1> <file or directory 2> <file or directory 3> ...
BaiduPCS-Go d <Netdisk file or directory path 1> <file or directory 2> <file or directory 3> ...
```

### Optional Parameters
```
  --test          Test download, this operation will not save files to local
  --ow            overwrite, overwrite existing files
  --status        Output working status of all threads
  --save          Save downloaded files directly to the current working directory
  --saveto value  Save downloaded files directly to the specified directory
  -x              Add execute permission to files, (invalid for windows system)
  --mode value    Download mode, optional values: pcs, stream, locate, default is locate, see help above for related instructions (default: "locate")
  -p value        Specify number of download threads (default: 0)
  -l value        Specify number of files downloading simultaneously (default: 0)
  --retry value   Maximum retry times for download failure (default: 3)
  --nocheck       Do not verify file after download is complete

```

Downloaded files are saved by default to the download/ directory of the **program's location**, supports setting a specified directory, files with the same name will be automatically skipped!

Downloaded files are saved by default to the **download/** directory of the **program's location**.

Use `BaiduPCS-Go config set -savedir <savedir>` to customize the save directory.

Supports downloading multiple files or directories.

Automatically skips downloading files with the same name!


#### Examples
```
# Set save directory, save to D:\Downloads
# Note the difference between backslash "\" and slash "/" !!!
BaiduPCS-Go config set -savedir D:/Downloads

# Download /My Resources/1.mp4
BaiduPCS-Go d /My Resources/1.mp4

# Download /My Resources entire directory!!
BaiduPCS-Go d /My Resources

# Download all files inside Netdisk!!
BaiduPCS-Go d /
BaiduPCS-Go d *
```

## Upload files/directories
```
BaiduPCS-Go upload <local file/directory path 1> <file/directory 2> <file/directory 3> ... <target directory>
BaiduPCS-Go u <local file/directory path 1> <file/directory 2> <file/directory 3> ... <target directory>
```

* Upload uses chunked upload by default, uploaded files will be saved to <target directory. Does not support resuming broken uploads.

* Files with the same name are automatically skipped, or you can configure `upload_policy` to choose to overwrite or only skip files of the same size.

* When the uploaded file name is the same as the Netdisk directory name, the directory will not be overwritten to prevent data loss.

* All uploads check for rapid upload by default, can add parameter `--norapid` to skip


#### Examples:
```
# Upload local C:\Users\Administrator\Desktop\1.mp4 to Netdisk /Video directory
# Note the difference between backslash "\" and slash "/" !!!
BaiduPCS-Go upload C:/Users/Administrator/Desktop/1.mp4 /Video

# Upload local C:\Users\Administrator\Desktop\1.mp4 to Netdisk /Video directory, do not check for rapid upload
BaiduPCS-Go upload C:/Users/Administrator/Desktop/1.mp4 /Video --norapid

# Upload local C:\Users\Administrator\Desktop\1.mp4 and C:\Users\Administrator\Desktop\2.mp4 to Netdisk /Video directory
BaiduPCS-Go upload C:/Users/Administrator/Desktop/1.mp4 C:/Users/Administrator/Desktop/2.mp4 /Video

# Upload local C:\Users\Administrator\Desktop entire directory to Netdisk /Video directory, only overwrite files with the same name that have different sizes from local
BaiduPCS-Go upload C:/Users/Administrator/Desktop /Video --policy rsync
```

## Get download direct link
```
BaiduPCS-Go locate <file 1> <file 2> ...
```

#### Examples:

```
BaiduPCS-Go config set -user_agent "netdisk;2.2.51.6;netdisk;10.0.63;PC;android-android"
```

## Export files/directories

```

BaiduPCS-Go export <file/directory 1> <file/directory 2> ...

BaiduPCS-Go ep <file/directory 1> <file/directory 2> ...

```

Export files or directories inside Netdisk, the principle is rapid upload of files, this operation will generate commands to export files or directories.

#### Note

**Rapid upload is no longer supported, this function has no actual effect**

#### Examples:

```

# Export current working directory:

BaiduPCS-Go export

# Export all files and directories, and set new root directory to /root

BaiduPCS-Go export -root=/root /

# Export /My Resources

BaiduPCS-Go export /My Resources

# Export /My Resources in generic rapid upload link format

BaiduPCS-Go export /My Resources --link

```

## Create directory
```
BaiduPCS-Go mkdir <directory>
```

#### Examples
```
BaiduPCS-Go mkdir 123
```

## Delete files/directories
```
BaiduPCS-Go rm <Netdisk file or directory path 1> <file or directory 2> <file or directory 3> ...
```

Note: When deleting multiple files and directories, please ensure every file and directory exists, otherwise the delete operation will fail.

Deleted files or directories can be retrieved in the Netdisk recycle bin.

#### Examples
```
# Delete /My Resources/1.mp4
BaiduPCS-Go rm /My Resources/1.mp4

# Delete /My Resources/1.mp4 and /My Resources/2.mp4
BaiduPCS-Go rm /My Resources/1.mp4 /My Resources/2.mp4

# Delete all files and directories inside /My Resources, but do not delete that directory
BaiduPCS-Go rm /My Resources/*

# Delete /My Resources entire directory !!
BaiduPCS-Go rm /My Resources
```

## Copy files/directories
```
BaiduPCS-Go cp <file/directory> <target file/directory>
BaiduPCS-Go cp <file/directory 1> <file/directory 2> <file/directory 3> ... <target directory>
```

Note: When copying multiple files and directories, please ensure every file and directory exists, otherwise the copy operation will fail.

#### Examples
```
# Copy /My Resources/1.mp4 to root directory /
BaiduPCS-Go cp /My Resources/1.mp4 /

# Copy /My Resources/1.mp4 and /My Resources/2.mp4 to root directory /
BaiduPCS-Go cp /My Resources/1.mp4 /My Resources/2.mp4 /
```

## Move/Rename files/directories
```
# Move:
BaiduPCS-Go mv <file/directory 1> <file/directory 2> <file/directory 3> ... <target directory>
# Rename:
BaiduPCS-Go mv <file/directory> <renamed file/directory>
```

Note: When moving multiple files and directories, please ensure every file and directory exists, otherwise the move operation will fail.

#### Examples
```
# Move /My Resources/1.mp4 to root directory /
BaiduPCS-Go mv /My Resources/1.mp4 /

# Rename /My Resources/1.mp4 to /My Resources/3.mp4
BaiduPCS-Go mv /My Resources/1.mp4 /My Resources/3.mp4
```

## Transfer files/directories
```
# Transfer files from shared link to current directory:
BaiduPCS-Go transfer <share link> <extraction code>
```

Note: Transferred files are saved to the current working directory, specifying a different one is not supported.

#### Examples
```
# Transfer https://pan.baidu.com/s/12L_ZZVNxz5f_2CccoyyVrW (extraction code edv4) to current directory
BaiduPCS-Go transfer https://pan.baidu.com/s/12L_ZZVNxz5f_2CccoyyVrW edv4
BaiduPCS-Go transfer https://pan.baidu.com/s/12L_ZZVNxz5f_2CccoyyVrW?pwd=edv4
```

## Share files/directories
```
BaiduPCS-Go share
```

### Set share for files/directories
```
BaiduPCS-Go share set <file/directory 1> <file/directory 2> ...
BaiduPCS-Go share s <file/directory 1> <file/directory 2> ...
```

### List shared files/directories
```
BaiduPCS-Go share list
BaiduPCS-Go share l
```

### Cancel share for files/directories
```
BaiduPCS-Go share cancel <shareid_1> <shareid_2> ...
BaiduPCS-Go share c <shareid_1> <shareid_2> ...
```

Currently only supports canceling sharing via share id (shareid).

## Offline Download
```
BaiduPCS-Go offlinedl
BaiduPCS-Go clouddl
BaiduPCS-Go od
```

Offline download supports http/https/ftp/eD2k/magnet link protocols.

There is a limit to the number of offline download tasks that can be performed simultaneously, parts exceeding the limit cannot be added.

### Add offline download task
```
BaiduPCS-Go offlinedl add -path=<path where offline download files are saved> resource address 1 address 2 ...
```

After successfully adding a task, the offline download task ID is returned.

### Query offline download task precisely
```
BaiduPCS-Go offlinedl query task ID 1 task ID 2 ...
```

### Query offline download task list
```
BaiduPCS-Go offlinedl list
```

### Cancel offline download task
```
BaiduPCS-Go offlinedl cancel task ID 1 task ID 2 ...
```

### Delete offline download task
```
BaiduPCS-Go offlinedl delete task ID 1 task ID 2 ...

# Clear offline download task records, the program will not ask for confirmation twice, operate with caution!!!
BaiduPCS-Go offlinedl delete -all
```

#### Examples
```
# Offline download Baidu and Tencent homepages to root directory /
BaiduPCS-Go offlinedl add -path=/ http://baidu.com http://qq.com

# Add magnet link task
BaiduPCS-Go offlinedl add magnet:?xt=urn:btih:xxx

# Query status of offline download task with ID 12345
BaiduPCS-Go offlinedl query 12345

# Cancel offline download task with ID 12345
BaiduPCS-Go offlinedl cancel 12345
```

## Recycle Bin
```
BaiduPCS-Go recycle
```

Recycle bin operations.

### List recycle bin files
```
BaiduPCS-Go recycle list
```

#### Optional Parameters
```
  --page value  Recycle bin file list page number (default: 1)
```

### Restore recycle bin files or directories
```
BaiduPCS-Go recycle restore <fs_id 1> <fs_id 2> <fs_id 3> ...
```

Restore specified files or directories from recycle bin based on file/directory fs_id.

### Delete recycle bin files or directories / Empty recycle bin
```
BaiduPCS-Go recycle delete [-all] <fs_id 1> <fs_id 2> <fs_id 3> ...
```

Delete specified files or directories from recycle bin or empty the recycle bin based on file/directory fs_id or -all parameter.

#### Examples
```
# Restore two files from recycle bin, the fs_id of the two files are 1013792297798440 and 643596340463870 respectively
BaiduPCS-Go recycle restore 1013792297798440 643596340463870

# Delete two files from recycle bin, the fs_id of the two files are 1013792297798440 and 643596340463870 respectively
BaiduPCS-Go recycle delete 1013792297798440 643596340463870

# Empty recycle bin, the program will not ask for confirmation twice, operate with caution!!!
BaiduPCS-Go recycle delete -all
```

## Display program environment variables
```
BaiduPCS-Go env
```

BAIDUPCS_GO_CONFIG_DIR: Configuration file path,

BAIDUPCS_GO_VERBOSE: Whether to enable debugging.

## Display and modify program configuration items
```
# Display configuration
BaiduPCS-Go config

# Set configuration
BaiduPCS-Go config set
```

Note: After v3.5, the program adjusted the search for the configuration file storage path. The directory where the configuration file is located can be the directory where the program itself is located, or the home directory.

Cases where the directory containing the configuration file is the home directory:

Windows: `%APPDATA%\BaiduPCS-Go`

Other operating systems: `$HOME/.config/BaiduPCS-Go`

You can specify the directory where configuration files are stored by setting the environment variable `BAIDUPCS_GO_CONFIG_DIR`.

Be cautious when modifying the values of `appid`, `user_agent`, `pcs_ua`, `pan_ua`, otherwise errors may occur when accessing the Netdisk server.

If you encounter exceptions during upload, you can try modifying `pcs_addr`. Currently known addresses are:

```
pcs.baidu.com
c.pcs.baidu.com
c2.pcs.baidu.com
c3.pcs.baidu.com
c4.pcs.baidu.com
c5.pcs.baidu.com
d.pcs.baidu.com
```
After v3.9.8, upload supports dynamically obtaining the pcs server, theoretically no need to configure manually. If you wish to use a static pcs server, you can open `fix_pcs_addr` in the configuration.

The value of `cache_size` supports optional unit setting, the unit is case-insensitive, both `b` and `B` mean bytes, e.g. `64KB`, `1MB`, `32kb`, `65536b`, `65536`.

The values of `max_download_rate`, `max_upload_rate` support optional unit setting, the unit is transmission rate per second, the suffix `/s` can be omitted, e.g. `2MB/s`, `2MB`, `2m`, `2mb` all mean the same thing.

**Normal users please set both `max_parallel` and `max_download_load` to 1. Increasing the thread count will only increase download speed for a short time, and is very likely to trigger speed limiting soon, causing the account to be near 0 speed on all clients for several hours to days. This software does not support speed acceleration for normal users.**

**SVIP users are recommended to set `max_parallel` to 10 or more, which can be increased according to actual bandwidth, but it is not recommended to exceed 20. Set `max_download_load` to 1 - 2. Experiments show that stable full-speed download can be achieved.**

#### Examples
```
# Display all settable values
BaiduPCS-Go config -h
BaiduPCS-Go config set -h

# Set storage directory for downloaded files
BaiduPCS-Go config set -savedir D:/Downloads

# Set maximum download concurrency to 15
BaiduPCS-Go config set -max_parallel 15

# Combined setting
BaiduPCS-Go config set -max_parallel 150 -savedir D:/Downloads
```

## Test wildcards
```
BaiduPCS-Go match <wildcard expression>
```

Test wildcard matching paths. If the operation is successful, all matched paths will be output.

#### Examples
```
# Match all mp4 format files under /My Resources directory
BaiduPCS-Go match /My Resources/*.mp4
```

## Toolbox
```
BaiduPCS-Go tool
```

Currently, the toolbox supports file encryption and decryption, etc.

# Beginner Tutorial

Suggestion for beginners: **Double-click to run the program**, entering the cli interactive mode imitating a Linux shell;

In cli interactive mode, the prefix of the line where the cursor is located should be `BaiduPCS-Go >`, if logged into a Baidu account the format is `BaiduPCS-Go:<working directory> <Baidu ID>$ `

The commands in the following examples are all commands in cli interactive mode.

Correct operation for running commands: **Type the command, press the Enter key (Enter on the keyboard)**, the program will receive the command and output the result.

## 1. View program usage instructions

In cli interactive mode, run the command `help`.

## 2. Login to Baidu Account (Required)

In cli interactive mode, run the command `login -h` (note the space) to view help.

In cli interactive mode, run the command `login`, the program will prompt you to enter your Baidu username (mobile number/email/username) and password, and if necessary, it can also perform online verification of the bound mobile number or email.

## 3. Switch Netdisk working directory

In cli interactive mode, run the command `cd /My Resources` to switch the working directory to `/My Resources` (Prerequisite: The directory exists in the Netdisk).

Directory supports wildcard matching, so you can also do this: run the command `cd /My*` or `cd /My??` to switch the working directory to `/My Resources`, simplifying input.

After successfully switching the working directory to `/My Resources`, run the command `cd ..` to switch to the parent directory, i.e., switching the working directory to `/`.

Why is it designed this way? For example,

Suppose you want to download two files named `1.mp4` and `2.mp4` inside `/My Resources`, and without switching the working directory, you need to run the following commands sequentially:

```
d /My Resources/1.mp4
d /My Resources/2.mp4
```

Whereafter switching the Netdisk working directory, run the following commands sequentially:

```
cd /My Resources
d 1.mp4
d 2.mp4
```

This achieves the purpose of simplifying input.

## 4. List files and directories in Netdisk

In cli interactive mode, run the command `ls -h` (note the space) to view help.

In cli interactive mode, run the command `ls` to list files and directories in the current directory.

In cli interactive mode, run the command `ls /My Resources` to list files and directories inside `/My Resources`.

In cli interactive mode, run the command `ls ..` to list files and directories in the parent directory of the current directory.

## 5. Download files

Description: Downloaded files are saved to the download/ directory (folder) by default.

In cli interactive mode, run the command `d -h` (note the space) to view help.

In cli interactive mode, run the command `d /My Resources/1.mp4` to download the file `1.mp4` located at `/My Resources/1.mp4`. This operation is equivalent to running the following commands:

```
cd /My Resources
d 1.mp4
```

Now directory (folder) download is supported, so, running the following command will download all files inside `/My Resources` (except for illegal files):

```
d /My Resources
```

## 6. Set maximum download concurrency

In cli interactive mode, run the command `config set -h` (note the space) to view setting help and values available for setting.

In cli interactive mode, run the command `config set -max_parallel 2` to set the maximum download concurrency to 2.

Note: Normal users setting the maximum download concurrency value over 1 will cause the account to be speed-limited; SVIP should also not set it too high, 10~20 is recommended.

## 7. Restore default configuration

In cli interactive mode, run the command `config reset`.

## 8. Exit program

Run the command `quit` or `exit` or key combination `Ctrl+C` or key combination `Ctrl+D`.

# Known Issues

* When uploading file in chunks, when the number of file chunks is greater than 1, the md5 value finally calculated by the Netdisk server is inconsistent with the local one. This may be a bug of Baidu Netdisk. Testing shows that after downloading the uploaded file to the local, the md5 values match. You can repair the md5 value using the principle of rapid upload.
* When MD5 verification download is enabled, there may be cases where check MD5 does not pass, but the file is actually not corrupted. Use --no-check download or enable no_check in the configuration (enabled by default in version 3.7).
* When logging in with username, the image captcha needs to be entered at least twice, the first input is invalid.
* When mobile/email verification appears during login, the image captcha needs to be entered at least 4 times.


# TODO

* Bypass the single limit for the number of transferred files

# Communication and Feedback

Submit Issue: [Issues](https://github.com/qjfoidnh/BaiduPCS-Go/issues)
