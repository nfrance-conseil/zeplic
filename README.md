zeplic
======

[![Build Status](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic.svg?branch=master)](https://travis-ci.org/IgnacioCarbajoVallejo/zeplic)

ZFS Datasets distribution over datacenter - Let'zeplic


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
