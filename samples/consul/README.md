## Consul datacenter configuration

If you want to use `zeplic --director`, the first step is to config your Consul datacenter:

### JSON config files

Sample file to config Consul with a server to execute `zeplic --director`:
- `/etc/consul.d/server/config.json` for Linux
- `/usr/local/etc/consul.d/server/config.json` for FreeBSD

```
{
	"bind_addr": "$IP_SERVER",
	"bootstrap_expect": 1,
	"server": true,
	"node_id": "00000000-aaaa-1111-9999-a0123456789z",
	"datacenter": "ConsulDatacenter",
	"node_name": "$NAME_SERVER",
	"data_dir": "/tmp/consul",
	"log_level": "DEBUG",
	"enable_syslog": true,
	"enable_debug": true,
	"disable_update_check": true,
	"leave_on_terminate": false,
	"skip_leave_on_interrupt": true,
	"rejoin_after_leave": true
}
```

Sample file to config Consul with several clients to run `zeplic --agent` and `zeplic --slave`:
- `/etc/consul.d/client/config.json` for Linux
- `/usr/local/etc/consul.d/client/config.json` for FreeBSD

```
{
	"bind_addr": "$IP_CLIENT",
	"server": false,
	"node_id": "00000000-aaaa-1111-9999-a0123456789z",
	"datacenter": "ConsulDatacenter",
	"node_name": "$NAME_CLIENT",
	"data_dir": "/tmp/consul",
	"log_level": "INFO", 
	"enable_syslog": true,
	"enable_debug": true,
	"disable_update_check": true,
	"check_update_interval": "1s",
	"leave_on_terminate": true,
	"skip_leave_on_interrupt": true,
	"rejoin_after_leave": true,
	"retry_join": ["$IP_SERVER"]
}
```

[Download Consul](https://www.consul.io/downloads.html)
