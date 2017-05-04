zeplic 0.1.0
============

[![Build Status](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic.svg?branch=master)](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**Tested on FreeBSD**

Process
-------

1. Get clone dataset
2. Destroy clone dataset
3. Get dataset (called in JSON file)
4. Destroy dataset (disable)
5. Create dataset if it does not exist
6. Save the last #Retain(JSON file) snapshots
7. Create a new snapshot
8. Create a clone of last snapshot
9. Rollback of last snapshot (disable)

Configuration
-------------

Use a JSON file (/etc/zeplic.d/config.json):

```sh
{
	"dataset": "tank/test",
	"clone": "tank/clone",
	"retain": 3
}
```

Utils
-----

- System logging daemon. Add the next line to your syslog configuration file ('/etc/syslog.conf'):

```sh
!zeplic
local0.*					/var/log/zeplic.log
```
