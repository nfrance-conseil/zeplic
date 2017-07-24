# zeplic v0.2.0-rc1

[![Build Status](https://travis-ci.org/nfrance-conseil/zeplic.svg?branch=master)](https://travis-ci.org/nfrance-conseil/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**zeplic is available for Linux and FreeBSD**

## Utils

1. Read your ZFS configuration from JSON file
2. Check your datasets enabled
3. Store log messages using syslog system service
4. Run ZFS functions...
- Destroy an existing clone (optional)
- Create dataset if it does not exist
- Create a new snapshot with an uuid
- Snapshots retention policy
- Create a backup snapshot (optional)
- Create a clone of last snapshot (optional)
5. Synchronisation between nodes using [Consul by HashiCorp](https://www.consul.io/)
- ZFS orders (OrderUUID, Action[take_snapshot, send_snapshot, destroy_snapshot], Destination, SnapshotUUID, SnapshotName, DestDataset, RollbackIfNeeded, SkipIfRenamed, SkipIfNotWritten, SkipIfCloned)
- Create a new snapshot
- Destroy a snapshot
- Send a snapshot via socket TCP

## How can you use it?

- First, clone this repository and type `(g)make`
- `sudo (g)make install` to install **zeplic**
- To clean all dependencies, type `(g)make clean`
- The next step is to configure **zeplic**:

### Configuration

You can modify a sample JSON file that it has been created in your config path:
- `/etc/zeplic/` for Linux
- `/usr/local/etc/zeplic/` for FreeBSD

```sh
{
	"local_datasets": [
	{
		"enable": true,
		"docker": false,
		"name": "tank/foo",
		"consul": {
			"enable": true,
			"datacenter": "ConsulDatacenter"
		},
		"snap_prefix": "FOO",
		"snap_retention": 24,
		"backup" true,
		"clone": {
			"enable": true,
			"name": "tank/foo_clone",
			"delete": true
		}
	},
	{
		"enable": false,
		"docker": false,
		"name": "tank/bar",
		...
	}]
}
```

- *enable*: to activate the dataset
- *docker*: dataset to receive the snapshots
- *name*: name of dataset
- *consul*: configuration using Consul (director's mode)
- *snap_prefix*: prefix of snapshot name (dataset@PREFIX_DATE)
- *snap_retention*: number of snapshots to save in local mode
- *backup*: backup snapshot of dataset (double copy)
- *clone*: make a clone of last snapshot created

### Running

**Let'zeplic!**

```sh
$ zeplic --run
```

Schedule a task with crontab to backup your files systems

```
MM	HH	*	*	*	root	$BINPATH/zeplic --run
```

### Director's mode
*In development...*

JSON file to configure the retention and replication policy. Use this one only in the server's node side:

```
{
	"datasets": [
	{
		"hostname": "localHostname",
		"datacenter": "ConsulDatacenter",
		"dataset": "tank/foo",
		"backup": {
			"creation": "00 * * * 1-5"
			"prefix": "BACKUP",
			"sync_on": "SyncHostname",
			"sync_dataset": "tank/copy_backup",
			"sync_policy": "00 1 * * *",
			"retention": "24d1w1m1y"
		"sync": {
			"creation": "00 4 * * *"
			"prefix": "SYNC",
			"sync_on": "SyncHostname",
			"sync_dataset": "tank/copy_sync",
			"sync_policy": "asap",
			"retention": "24d1w1m1y"
		},
		"rollback_needed": true,
		"skip_renamed": true,
		"skip_not_written": true,
		"skip_cloned": true
	}]
}
```

- *hostname*: hostname of local node
- *datacenter*: datacenter of Consul
- *dataset*: name of dataset to manage
- *creation*: policy to create a new snapshot (cron)
- *prefix*: prefix of snapshot name
- *sync_on*: node to synchronize
- *sync_dataset*: dataset in slave node
- *sync_policy*: policy to synchronize (asap | cron)
- *retention*: policy of retention
- other options

Formats for creation, send and destroy a snapshot:

```
Create: cron
Send: asap (as soon as possible) or cron
Destroy: DdWwMmYy
	- D = number of snapshots to retain in last 24h
	- W = number of snapshots to retain in last week
	- M = number of snapshots to retain in last month
	- Y = number of snapshots to retain in last year

Cron format = MM HH Monthday Month Weekday
```

- Send an order to the agent node (zeplic --agent) on port 7711 to create a snapshot or destroy it
- Send a snapshot between from agent's node (zeplic --agent) to slave's node (zeplic --slave)

### Syslog system service

Configure **zeplic** to send log messages to local/remote syslog server:
- Information of snapshots created, deleted, cloned...
- Errors occurred while running **zeplic**
- Information of the synchronisation between nodes

```
{
	"enable": true,
	"mode": "local",
	"info": "LOCAL0"
}
```
- *info(local): facility [LOCAL0-7]*
- *info(remote): tcp/upd:IP:port*

```
Jun 28 10:30:00 hostname zeplic[1364]: [INFO] the snapshot 'tank/foo@FOO_2017-June-28_10:00:00' has been sent.
Jun 29 10:00:00 hostname zeplic[1176]: [INFO] the snapshot 'tank/foo@FOO_2017-June-29_10:00:00' has been created.
Jun 29 10:00:00 hostname zeplic[1176]: [INFO] the snapshot 'tank/foo@BACKUP_from_2017-June-28' has been destroyed.
Jun 29 10:00:00 hostname zeplic[1176]: [WARNING] the snapshot 'tank/foo@FOO_2017-June-23_10:00:00' has dependent clones: 'tank/test_clone'.
Jun 29 10:00:00 hostname zeplic[1176]: [INFO] the snapshot 'tank/foo@FOO_2017-June-24_10:00:00' has been destroyed.
Jun 29 10:00:00 hostname zeplic[1176]: [INFO] the backup snapshot 'tank/foo@BACKUP_from_2017-June-25' has been created.
Jun 29 10:00:00 hostname zeplic[1176]: [INFO] the snapshot 'tank/foo@FOO_2017-June-29_10:00:00' has been clone.
Jun 29 10:00:00 hostname zeplic[1176]: [NOTICE] the dataset 'tank/bar' is disabled.
```

### Help menu

```
$ zeplic --help
Usage: zeplic [-acdrsv] [--help] [--quit] [parameters ...]
 -a, --agent     Execute the orders from director
 -c, --cleaner   Clean KV pairs with #deleted flag
 -d, --director  Execute 'zeplic' in synchronization mode
     --help      Show help menu
     --quit      Gracefully shutdown
 -r, --run       Execute 'zeplic' in local mode
 -s, --slave     Receive a new snapshot from agent
 -v, --version   Show version of zeplic
```

### Vendoring
**zeplic** currently uses [govendor](https://github.com/kardianos/govendor) for vendoring

### Version
**zeplic** uses [Semantic Versioning](http://semver.org/)
