# monitor

Basic monitoring system with multiple nodes.

## TODOs

* node: implement `reboot_required` (easy)
* Report(): make format consistent (json?)
* webui: do not log -vvv into the website
  * parse json in frontend and print only some/partial reports
  * do not send all reports to website?
* webui: keep state which reports are shown/relevant (e.g. send refresh event
  from server to client)

## Architecture

Nodes:
- have almost no logic
- start up, send data to the daemon (single json object) and terminate
- connect autonomously to the daemon

Daemon:
- listens continuously for new node messages
- sends reports to configured "reporters" if something needs attention

Reporter:
a daemon "output plugin" which sends the information (e.g. mail, redmine posts).

Collector:
a daemon "input plugin" which reads new information, e.g. node information via
"listener collector", or certificate expiry dates via "cert collector".

## Features

Don't:
- verify or authenticate nodes, anything can send messages
- track or monitor memory or cpu usage
- check for processes or services

Report when:
- Disk usage is beyond a certain threshold
- A reboot is required after apt upgrade
- A node stopped sending messages
- Certificates nearing end of validity

Try to:
- Have no external dependencies
- Nodes without configuration
- Nodes always report all information

## Example node json

```
{
  "host_name": "footop",
  "operating_system_name": "linux",
  "operating_system_version": "6.12.91",
  "reboot_required": false,
  "file_systems": [
    {
      "source": "/dev/mapper/foo--vg--extern-root",
      "used_bytes": 731363332,
      "available_bytes": 163412204,
      "capacity": "82%",
      "mount_point": "/"
    },
    {
      "source": "/dev/nvme0n1p1",
      "used_bytes": 178512,
      "available_bytes": 868000,
      "capacity": "18%",
      "mount_point": "/boot"
    }
  ]
}
```
