# zeplic v0.3.8

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
- `/etc/zeplic/local.json` for Linux
- `/usr/local/etc/zeplic/local.json` for FreeBSD

```sh
{
	"local_datasets": [
	{
		"enable": true,
		"slave": false,
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
- *slave*: dataset to receive the snapshots
- *name*: name of dataset
- *consul*: configuration using Consul (director's mode)
- *snap_prefix*: prefix of snapshot name (dataset@PREFIX_DATE)
- *snap_retention*: number of snapshots to save in local mode
- *backup*: backup snapshot of dataset (double copy)
- *clone*: make a clone of last snapshot created

### Local running

**Let'zeplic!**

```sh
$ zeplic --run
```

Schedule a task with crontab to backup your files systems

```
M	H	mday	month	wday	root	$BINPATH/zeplic --run
```

### Director's mode

JSON file to configure the retention and replication policy. Use this one only in the server's node side:
- `/etc/zeplic/server.json` for Linux
- `/usr/local/etc/zeplic/server.json` for FreeBSD

```
{
	"datacenter": "ConsulDatacenter",
	"consul_resync": ["19:00", "19:10"],
	"datasets": [
	{
		"hostname": "localHostname",
		"dataset": "tank/foo",
		"backup": {
			"creation": "0 * * * 1-5",
			"prefix": "BACKUP",
			"sync_on": "SyncHostname",
			"sync_dataset": "tank/copy_backup",
			"sync_policy": "0 1 * * *",
			"retention": ["24 in last day", "1/day in last week", "1/week in last month", "1/month in last year"]
		},
		"sync": {
			"creation": "0 4 * * *",
			"prefix": "SYNC",
			"sync_on": "SyncHostname",
			"sync_dataset": "tank/copy_sync",
			"sync_policy": "asap",
			"retention": ["24 in last day", "1/day in last week", "1/week in last month", "1/month in last year"]
		},
		"rollback_needed": true,
		"skip_renamed": true,
		"skip_not_written": true,
		"skip_cloned": true
	}]
}
```

- *datacenter*: datacenter of Consul
- *consul_resync*: time to resynchronize Consul data
- *hostname*: hostname of local node
- *dataset*: name of dataset to manage
- *creation*: policy to create a new snapshot (cron)
- *prefix*: prefix of snapshot name
- *sync_on*: node to synchronize
- *sync_dataset*: dataset in slave node
- *sync_policy*: policy to synchronize (asap | cron)
- *retention*: policy of retention
- *rollback_needed*: to do roll back if it is necessary
- *skip_renamed*: skip if the snapshot was renamed
- *skip_not_written*: skip if nothing new was written
- *skip_cloned*: skip if the snapshot was cloned

Formats for creation, send and destroy a snapshot:

```
Create: cron format
Send: asap (as soon as possible) or cron format
Destroy: ["D in last day", "W/day in last week", "M/week in last month", "Y/month in last year"]
	- D = snapshots to save in last 24h
	- W = snapshots to save per day in the last week
	- M = snapshots to save per week in the last month
	- Y = snapshots to save per month in the last year
```

- Send an order to the agent node (zeplic --agent) on port 7711 to create a snapshot or destroy it
- Send a snapshot between from agent's node (zeplic --agent) to slave's node (zeplic --slave)

### Consul configuration

How can you config your [Consul datacenter](https://github.com/nfrance-conseil/zeplic/tree/master/samples/consul)?

### Syslog system service

Configure **zeplic** to send log messages to local/remote syslog server:
- `/etc/zeplic/server.json` for Linux
- `/usr/local/etc/zeplic/server.json` for FreeBSD
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

- Syslog file format:

```
	- mode: local;    info: *facility [LOCAL0-7]*
	- mode: remote;   info: *tcp/upd:IP:port*
```

- Sample:

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

## Help menu

```
$ zeplic --help
Usage: zeplic [-acdrsv] [--help] [--quit] [parameters ...]
 -a, --agent     Execute the orders from director
 -c, --cleaner   Clean KV pairs with #deleted flag in a dataset
 -d, --director  Execute 'zeplic' in synchronization mode
     --help      Show help menu
     --quit      Gracefully shutdown
 -r, --run       Execute 'zeplic' in local mode
 -s, --slave     Receive a new snapshot from agent
 -v, --version   Show version of zeplic
```

## Vendoring
**zeplic** currently uses [govendor](https://github.com/kardianos/govendor) for vendoring

## Version
**zeplic** uses [Semantic Versioning](http://semver.org/)
