# monitor

Basic monitoring system of one daemon and multiple nodes.

## Daemon Architecture

To allow a dynamic configuration of different reports or data
collections the daemon is split up into subsystems.
On startup the daemon loads a configuration file which is
globally available as `Cfg` (see `config/config.go`).

It then starts all configured subsystems. For a list of all available subsystems
see `AvailableSubsystems` in `cmd/daemon/main.go`. The subsystem interface
requires only the method `Init() error` (see `subsystems/subsystem.go`)
which should start its goroutine and return.

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
Anything can publish messages on the bus, which will get delivered to
everyone that subscribed.

### Example subsystems

Let's say you have a subsystem that generates reports on disk usage
every hour. All it has to do is call `Publish` on the bus with the
information about the disks:

```go
type DiskUsage struct {
	Path  string // e.g. "/" or "/mnt"
	Usage int    // Usage in %, between 0 and 100
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
sends an all-clear message (here `DiskFineAgain`) if the disk's usage
is below the threshold again:

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
    				// Disk is getting full and we have not sent a report yet
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
threshold messages by printing them to `stdout`:

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

### Bus and message interfaces

There are all kinds of messages on the bus, some are just for
communication between two subsystems, some might be interesting for
the user and some are important messages that should send a notification
to the user.

Besides multiple subsystems that generate information like the node
listener or the certificate check, there might also be
subsystems that report messages to the user, e.g. the stdout or the
pushover subsystem that sends notifications.

They won't print out any message they receive from the bus, since
not all messages are relevant. That's why there are multiple
message interfaces, to be able to figure out which messages should get
reported. Any message that is relevant for the user implements the
`Info interface` (see `bus/bus.go`). Messages with a higher importance
also implement the `Important interface`.
This way a newly added subsystem can report new messages that should get
 reported without having to update all reporting subsystems.

Having two interfaces allows for filtering, e.g. the pushover
subsystems is only interested in important messages while a logging
subsystem would log any information.

### Critical Info messages

Most important info messages are critical and they should persist until
they are resolved. For example if a cert is nearing end of life, the message
is relevant until the certificate is renewed.
This is implemented as critical messages. The store (see
`store/store.go`) will listen for every Info message and keep critical
messages until a non-critical Info message with the same ID is
published.
This allows one to get all critical messages that occured but were not
cleared yet (see `store.FetchCritical`).

## Node Architecture

Nodes are other hosts that report autonomously, regularly and
unencrypted to the daemon via TCP.

Nodes run a simple binary (also called `node`), which collects
and sends the JSON formatted information via TCP to the daemon.

The daemon's address is passed via command line arguments.

### Example node JSON

```
{
  "hostname": "examplehost",
  "operating_system_name": "linux",
  "operating_system_version": "6.12.91",
  "reboot_required": false,
  "file_systems": [
    {
      "source": "/dev/mapper/example--vg-root",
      "used_bytes": 710166144,
      "available_bytes": 184609392,
      "capacity": "80%",
      "mount_point": "/"
    },
    {
      "source": "/dev/nvme0n1p1",
      "used_bytes": 127952,
      "available_bytes": 918560,
      "capacity": "13%",
      "mount_point": "/boot"
    }
  ]
}
```

## Third-Party dependencies

Take a look into `go.mod`:
So far, we have managed to avoid relying on any third-party code dependencies.
