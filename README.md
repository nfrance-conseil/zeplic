# zeplic

[![Build Status](https://travis-ci.org/nfrance-conseil/zeplic.svg?branch=master)](https://travis-ci.org/nfrance-conseil/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**zeplic is available for Linux and FreeBSD**

## Utils

1. Read your ZFS configuration from JSON file
2. Check your datasets enabled
3. Store log messages using syslog system service
4. Run ZFS functions...
- Destroy an existing clone
- Select datasets
- Destroy dataset (disable)
- Create dataset if it does not exist
- Create a new snapshot with an uuid
- Snapshots retention policy
- Create a backup snapshot (optional function)
- Create a clone of last snapshot (optional function)
- Rollback of last snapshot (optional function)
5. *In development...* Synchronisation between nodes using [Consul by HashiCorp](https://www.consul.io/)
- ZFS orders (order with UUID, action, snapshot uuid, NotWritten, Rollback, Renamed...)
6. *In development...* **zeplic** runs as background

## How can you use it?

- First, clone this repository and type `make|gmake` 
- After, type `sudo make|gmake install` to install **zeplic**
- If you want, you can clean all dependencies with `make|gmake clean`.
- The next step is to configure **zeplic**:

### Configuration

You can modify a sample JSON file that it has been created in your config path:
- `/etc/zeplic/` for Linux
- `/usr/local/etc/zeplic/` for FreeBSD

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

### Syslog system service

Check all actions of **zeplic** in:
```
$ /var/log/zeplic.log
```
- Information of snapshots created, deleted, cloned...
- Errors occurred while running **zeplic**
- *In development...* Information of the synchronisation between nodes

### Daemon service

*In development...* **zeplic** runs as background
