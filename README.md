# zeplic

[![Build Status](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic.svg?branch=master)](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**Tested on FreeBSD**

## Utils

1. Run syslog service
2. Read JSON configuration file
3. Check datasets enabled
4. ZFS functions...
  4.1. Destroy an existing clone
  4.2. Select datasets
  4.3. Destroy dataset (disable)
  4.4. Create dataset if it does not exist
  4.5. Create a new snapshot
  4.6. Save the last #Retain(JSON file) snapshots
  4.7. Create a backup snapshot
  4.8. Create a clone of last snapshot (optional function)
  4.9. Rollback of last snapshot (optional function)

## How can you use it?

First, clone this repository into `$GOPATH/src/zeplic` and export your `$GOBIN`.
After, make `go install`.
The next step is to configure **zeplic**:

### Configuration

Add the next line to your syslog configuration file `/etc/syslog.conf`:

```sh
!zeplic
local0.*					-/var/log/zeplic.log
```

Use a JSON file (/usr/local/etc/zeplic.d/config.json):

```sh
{
	"dataset": [
		"enable": true,
		"name": "tank/test",
		"snapshot": "SNAP",
		"retain": 5,
		"backup" true,
		"clone": {
			"enable": true,
			"tank/clone"
		},
		"rollback": false
	},
	{
		"enable": false,
		"name": "tank/storage"
		...
	}]
}
```

Finally, **let'zeplic!**

```sh
$ zeplic
```
