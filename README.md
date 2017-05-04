# zeplic 0.1.0

[![Build Status](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic.svg?branch=master)](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**Tested on FreeBSD**

## Process

1. Get clone dataset
2. Destroy clone dataset
3. Get dataset (called in JSON file)
4. Destroy dataset (disable)
5. Create dataset if it does not exist
6. Save the last #Retain(JSON file) snapshots
7. Create a new snapshot
8. Create a clone of last snapshot
9. Rollback of last snapshot (disable)

## Utils

- System logging daemon.

## How can you use it?

First, clone this repository into `$GOPATH/src/github.com/nfrance-conseil/zeplic` and export your `$GOBIN`.
The next step is to configure **zeplic**:

### Configuration

Add the next line to your syslog configuration file `/etc/syslog.conf`:

```sh
!zeplic
local0.*					/var/log/zeplic.d
```

Use a JSON file (/etc/zeplic.d/config.json):

```sh
{
	"dataset": "tank/test",
	"clone": "tank/clone",
	"retain": 3
}
```

Finally, **let'zeplic!**

```sh
$ zeplic
```
