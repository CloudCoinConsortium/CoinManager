# SuperRAIDA Client Backend

![SuperRAIDA]()

SuperRAIDA Client Backend is a backend part of the UI Client for the SuperRAIDA. It is writtend in Go Languange and provides RESTful interface to process CloudCoins

## Installing SuperRAIDA Client Backend

1. Copy the superraida binary (Linux) or superraida.exe binary (Windows) to any location (e.g. /usr/local/bin for Linux)

2. Copy files from the Assets folder to the Templates folder (/home/user/superraida/Templates or c:\Users\User\superraida\Templates)

[-version](README.md#-version)

[-help](README.md#-help)

[-debug](README.md#-debug)

[Config](README.md#config)


## -version
example usage:
```
C:\cloudcoin\superraida.exe -version
```
Sample response:
```
0.0.3
```

## -help

example usage:
```
C:\cloudcoin\superraida.exe -help
```
Sample response:
```console
Usage of superraida:
superraida [-debug] [-log logfile] <operation> <args>
superraida [-help]
superraida [-version]

<args> arguments for operation

  -debug
        Display Debug Information
  -help
        Show Usage
  -version
        Display version
```

## Config
You can configure the behaviour of the RaidaGo by using a configuration file. The file must be placed in your superraida folder. This folder is located in the user's directory.

For Linux:
/home/user/superraida/config.toml

For Windows:
c:\Users\User\superraida\config.toml

The file is in TOML format (https://en.wikipedia.org/wiki/TOML):
Only three directives are supported at the moment

```toml
title = "SuperRAIDA Client Backend Configuration File"
  
[main]
# connection and read timeout
timeout = 40

# main domain name
main_domain = "cloudcoin.global"

# max number of notes that can be trasferred at a time
max_fixtransfer_notes = 400

