# zeplic

[![Build Status](https://travis-ci.org/nfrance-conseil/zeplic.svg?branch=master)](https://travis-ci.org/nfrance-conseil/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**Tested on FreeBSD**

## Utils

1. Run syslog service
2. Read JSON configuration file
3. Check datasets enabled
4. ZFS functions...
- Destroy an existing clone
- Select datasets
- Destroy dataset (disable)
- Create dataset if it does not exist
- Create a new snapshot
- Save the last #Retain(JSON file) snapshots
- Create a backup snapshot
- Create a clone of last snapshot (optional function)
- Rollback of last snapshot (optional function)

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
