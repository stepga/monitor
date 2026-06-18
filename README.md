# monitor

Basic monitoring system with multiple nodes.

## TODOs

* Info(): make format consistent (json?)
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
up all configured subsystems. For a list of all available subsystems
see `AvailableSubsystems` in `cmd/daemon/main.go`. The subsystem
interface requires only a single Init method (`Init() error`) (see
`subsystems/subsystem.go`) which should start its goroutines and
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

Anything can publish messages on the bus which will get delivered to
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

These events can now be tracked by another subsystem that sends
errors if a disk has a usage over 80% (here `DiskGettingFull`) or
sends an ok message (here `DiskFineAgain`) if the disk is below the
threshold again:

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

Now we could implement a third subsystem that simply reports all disk
threshold messages by printing them on stdout:


```go
type StdoutReporter struct{}

func (_ *StdoutReporter) Init() error {
	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for m := range ch {
			switch msg := m.(type) {
			case DiskGettingFull:
				fmt.Printf("Warning: disk %s is getting full: %d%%!\n", msg.Path, msg.Usage)
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

And add them to the `subsystems` array in your `config.json`:

```json
   "subsystems": ["DiskUsageReporter", "DiskUsageTracker", "StdoutReporter"],
```

## Bus and message interfaces

There are all kinds of messages on the bus, some are just for
communication between two subsystems, some might be interesting for
the user and some are important messages that should send a
notification to the user.
Besides multiple subsystems that generate information like the node
listener or the certificate check, there might also be
subsystems that report messages to the user, e.g. the stdout or the
pushover subsystem that sends notifications.
They cant just print out any message they receive from the bus, since
not all messages are relevant. Thats why there are multiple
interfaces for them to be able to figure out which messages should get
reportet. Any message that is relevant for the user implements the
Info interface (see `bus/bus.go`). Messages with a higher importance
also implement the Important interace.
This way a new subystem can report new messages that should get
reportet without having to update all reporting subsystems.
Having two interfaces allows for filtering, e.g. the pushover
subystems is only interested in important messages while a logging
subsystem would log any information.

## Sticky Info messages

Most important info messages are sticky, in the sense that they should
"stick around" until they are resolved. For example if a cert is
nearing end of life, the message is relevant until the certificate is
renewed.
This is implemented as sticky messages. The store (see
`store/store.go`) will listen for every Info message and keep sticky
messages until a non-sticky Info message with the same ID is
published.
This allows one to get all sticky messages that occured but were not
cleared yet (see `store.FetchSticky`).
