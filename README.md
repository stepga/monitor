# monitor

Basic monitoring system with multiple nodes.

## TODOs

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
  "hostname": "footop",
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

# Code

## Daemon Architecture

To allow a dynamic configuration of different reportings or data
collections the daemon is split up into subsystems (TODO better
name?). On startup the daemon loads a configuration file which is
globally available as `Cfg` (see `config/config.go`). It then starts
up all configured subsystems. For all available subsystems, see
`AvailableSubsystems` in `cmd/daemon/main.go`. The subsystem interface
is straight forward, it only defines a single method called `Init`
(see `subsystems/subsystem.go`) which should start its goroutines and
return.
Subsystems talk to each other through a global bus (see `bus/bus.go`).
Some subsystems only generate messages (e.g. `subsystem/cert.go` or
`subsystem/listener.go`) and others might only consume some of the
messages (e.g. `subsystem/stdout.go`). They might also do both.

```
┌──────────────────────┐
│       subsystem      │
│──────────────────────│
│ Publish(any) ────────│────┐
│                      │    │
│ Subscribe() ─────────│──┐ │
│ returns channel      │  │ │
│ ┌──────────────────┐ │  │ │
│ │     channel      │ │  │ │
│ └─────────▲────────┘ │  │ │
└───────────│──────────┘  │ │
     msg    │             │ │
   delivery │             │ │
            │             │ │
            │    ┌────────▼─▼────────┐
            └────┤        bus        │
                 │───────────────────│
                 │ Publish(any)      │
                 │ Subscribe() chan  │
                 │ Unsubscribe(chan) │
                 └───────────────────┘
```
Anything can publish messages on the bus, they will get delivered to
everyone that subscribed.

## Example subsystems

Lets say you have a subsystem that generates reports on disk usage
every hour. All it has to do is call `Publish` on the bus with the
information about the disks:

```go
type DiskUsage struct {
	Path  string // e.g. "/" or "/mnt"
	Usage int    // Usage in %, between 0 und 100
}

type DiskUsageReporter struct{}

// Implement subsystem interface
func (_ *DiskUsageReporter) Init() error {
	go func() {
		for {
			// get disk info somehow
			root := DiskUsage{
				Path:  "/",
				Usage: 81,
			}
			bus.Publish(root)
			time.Sleep(1 * time.Hour)
		}
	}()
    return nil
}
```

These events could now be tracked by a tracker, that sends errors if a
disk has a usage over 80% (here `DiskGettingFull`) and sends an ok
message (here `DiskFineAgain`) if the disk is below the threshold
again:

```go
type DiskUsageTracker struct{}

type DiskGettingFull struct {
	DiskUsage
}

type DiskFineAgain struct {
	DiskUsage
}

func (_ *DiskUsageTracker) Init() error {
    go func() {
    	ch := bus.Subscribe()
    	defer bus.Unsubscribe(ch)
    	// Track which disk we already reported
    	reported := make(map[string]struct{})
    	for m := range ch {
    		switch msg := m.(type) {
    		case DiskUsage:
    			_, exists := reported[msg.Path]
    			if msg.Usage > 80 && !exists {
    				// Disk is getting full and we havent
    				// sent a report yet
    				gettingFull := DiskGettingFull{}
    				gettingFull.Path = msg.Path
    				gettingFull.Usage = msg.Usage
    				bus.Publish(gettingFull)
    				reported[msg.Path] = struct{}{}
    			} else if msg.Usage < 80 && exists {
    				// Disk where we have sent a report is
    				// below threshold again
    				fineAgain := DiskFineAgain{}
    				fineAgain.Path = msg.Path
    				fineAgain.Usage = msg.Usage
    				bus.Publish(fineAgain)
    				delete(reported, msg.Path)
    			}
    		}
    	}
    }()
    return nil
}
```

Now we could implement a third subsystem, that simply reports all disk
messages reported by the tracker by printing them on stdout:


```go
type StdoutReporter struct{}

func (_ *StdoutReporter) Init() error {
	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for m := range ch {
			switch msg := m.(type) {
			case DiskGettingFull:
				fmt.Printf("Warning, disk %s is getting full: %d%%!\n", msg.Path, msg.Usage)
			case DiskFineAgain:
				fmt.Printf("Disk %s is fine again: %d%%.\n", msg.Path, msg.Usage)
			}

		}
	}()
    return nil
}
```

To be able to start and stop these from the config file (see
`subsystems` in `config/config.go`), simply add them to
`AvailableSubsystems` in `cmd/daemon/main.go`:
 
```go
var AvailableSubsystems = map[string]subsystems.Subsystem{
	"DiskUsageReporter": &DiskUsageReporter{},
	"DiskUsageTracker":  &DiskUsageTracker{},
	"StdoutReporter":    &StdoutReporter{},
}
```

And add them to the `subsystems` in your `config.json`:
```json
   "subsystems": ["DiskUsageReporter", "DiskUsageTracker", "StdoutReporter"],
```

## Bus and message interfaces

There are all kinds of messages on the bus, some are just for
communication between two subsystems, some might be interesting for
the user and some are important messages that should notify the user.
Besides multiple subsystems that generated information like the node
listener or the certificate check, there might also be multiple
subsystems that report messages to the user, e.g. the stdout reporter
or the pushover subsystem that sends notifications. These must be able
to figure out which messages should get reportet. To avoid a big
switch case in every reporter that filters for the relevant messages,
all relevant messages implement the Info Interface and important
messages also implement the Important interface (see `bus/bus.go`).
This way a new subystem can report new messages that should get
reportet without having to update all reporting subsystems.
Having two interfaces allows for filtering, e.g. the pushover
subystems is only interested in important messages while a logging
subsystem would log any information.
