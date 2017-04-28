zeplic 0.1.0
============

[![Build Status](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic.svg?branch=master)](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic


Process
-------

1. Get clones dataset
2. Destroy clones dataset
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
	"clones": "tank/clones",
	"retain": 3
}
```
