# monitor

Basic monitoring system with multiple nodes that connect to a daemon
that sends reports if something needs attention

## Architecture

Almost no logic on nodes. They start up, report to the daemon
(single json object) and terminate

## Features

Dont:
- verify or authenticate Nodes, anything can report
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
  "name": "foo.bar.de",
  "os": "linux",
  "version": "Ubuntu 24.04.4 LTS"
  "disks": [
    ["/boot", 178512, 1046512],
    ["/mount/foo", 123213, 2321312]
  ],
  "apt_reboot_required": true
}
```
