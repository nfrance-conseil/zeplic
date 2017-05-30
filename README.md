# zeplic

[![Build Status](https://travis-ci.org/nfrance-conseil/zeplic.svg?branch=master)](https://travis-ci.org/nfrance-conseil/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**zeplic is available for FreeBSD**

## Utils

1. Run syslog service
2. Read JSON configuration file
3. Check datasets enabled
4. Run ZFS functions...
- Destroy an existing clone
- Select datasets
- Destroy dataset (disable)
- Create dataset if it does not exist
- Create a new snapshot
- Snapshots retention policy
- Create a backup snapshot
- Create a clone of last snapshot (optional function)
- Rollback of last snapshot (optional function)

## How can you use it?

- First, clone this repository and type `gmake` 
- After, type `sudo gmake install` to install **zeplic** and if you want, you can clean all dependencies with `gmake clean`.
- The next step is to configure **zeplic**:

### Configuration

Use a JSON file (/usr/local/etc/zeplic.d/config.json):

```sh
{
	"datasets": [
	{
		"enable": true,
		"name": "tank/test",
		"snapshot": "SNAP",
		"retain": 5,
		"backup" true,
		"clone": {
			"enable": true,
			"name": "tank/clone"
		},
		"rollback": false
	},
	{
		"enable": false,
		"name": "tank/storage",
		...
	}]
}
```

### Running

**Let'zeplic!**

```sh
$ zeplic
```
