# zeplic

[![Build Status](https://travis-ci.org/nfrance-conseil/zeplic.svg?branch=master)](https://travis-ci.org/nfrance-conseil/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic

**zeplic is available for Linux and BSD os**

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
- Send a snapshot or incremental stream via socket TCP (director mode)
- Snapshots retention policy
- Destroy a snapshot (director mode)
- Create a backup snapshot (optional function)
- Create a clone of last snapshot (optional function)
- Rollback of snapshot (director mode)
5. *In development...* Synchronisation between nodes using [Consul by HashiCorp](https://www.consul.io/)
- ZFS orders (OrderUUID, Action[take_snapshot, send_snapshot, destroy_snapshot], Destination, Snapshot UUID, RollbackIfNeeded, SkipIfRenamed, SkipIfNotWritten)

## How can you use it?

- First, clone this repository and type `make|gmake` 
- After, type `sudo make|gmake install` to install **zeplic**
- To clean all dependencies, type `make|gmake clean`.
- The next step is to configure **zeplic**:

### Configuration

You can modify a sample JSON file that it has been created in your config path:
- `/etc/zeplic/` for Linux
- `/usr/local/etc/zeplic/` for BSD os

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
		}
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
$ zeplic --run
```

### Director mode
*In development...*

You can send an order to the agent node (zeplic --agent) on port 7711:
- Create a snapshot
- Destroy a snapshot

```
$ echo '{"OrderUUID":"4fa34d08-51a6-11e7-a181-b18db42d304e","Action":"take_snapshot","Destination":"","SnapshotUUID":"","SnapshotName":"","DestDataset":"$DATASET_OF_SNAPSHOT","RollbackIfNeeded":false,"SkipIfRenamed":false,"SkipIfNotWritten":false}' | nc -w 3 $IP_AGENT 7711

$ echo '{"OrderUUID":"4fa34d08-51a6-11e7-a181-b18db42d304e","Action":"destroy_snapshot","Destination":"","SnapshotUUID":"$UUID_OF_SNAPSHOT","SnapshotName":"$NAME_OF_SNAPSHOT","DestDataset":"","RollbackIfNeeded":false,"SkipIfRenamed":false,"SkipIfNotWritten":false}' | nc -w 3 $IP_AGENT 7711
```

You can send a snapshot between the agent node (zeplic --agent) to the slave node (zeplic --slave):

```
echo '{"OrderUUID":"4fa34d08-51a6-11e7-a181-b18db42d304e","Action":"send_snapshot","Destination":"$HOSTNAME_SLAVE","SnapshotUUID":"$UUID_OF_SNAPSHOT","SnapshotName":"","DestDataset":"$DATASET_OF_DESTINATION",RollbackIfNeeded":false,"SkipIfRenamed":false,"SkipIfNotWritten":false}' | nc -w 3 $IP_AGENT 7711
```

### Syslog system service

Configure **zeplic** to send log messages to local/remote syslog server:
- Information of snapshots created, deleted, cloned...
- Errors occurred while running **zeplic**
- Information of the synchronisation between nodes

### Help menu

```
$ zeplic --help
Usage: zeplic [-adrsv] [--help] [--quit] [parameters ...]
 -a, --agent     Listen ZFS orders from director
 -d, --director  Send ZFS orders to agent
     --help      Help
     --quit      Gracefully shutdown
 -r, --run       Execute ZFS functions
 -s, --slave     Receive a new snapshot from agent
 -v, --version   Show version of zeplic

```
