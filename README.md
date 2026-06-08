# monitor

h3. architektur

- Ein großer daemon und @n@ kleine Nodes (eine Node per Maschine)
- Wenig bis keine Logik auf den Nodes
- Nodes verbinden sich zu Daemon

h3. Non-Goals

- Content von den Nodes wird nicht verifiziert
- Nodes authentisieren sich nicht
- Nodes überwachen keine Prozesse
- Nodes überwachen keine Memory Auslastung

h3. Features

- Disk usage
- apt upgrade status => reboot noetig (z.b. nach kernel update, oder initramfs-update) (@/var/run/reboot-required@)
- So wenig (am besten garkeine) dependencies wie moeglich
- Heartbeat
- Laufzeit Zertifikate: Daemon hat liste an https urls die er täglich checkt
- Node hat keine Konfiguration
- Node schickt immer alles an info an den Daemon
- Logik wie z.B. disks die ignoriert werden ist im Daemon

spaeter:
- cron job fails

h3. Implementierungs Detail

Daemon config
```
{
 "hosts": [
   "https://docs.foo.bar",
   "https://jenkins.org.name"
 ]
}
```

Node -> Daemon json, einmal in der Stunde
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

Daemon meldet Fehler wenn:

- Kein Heartbeat (2 missing oder so?)
- Disk > 80% (oder so)
- Reboot required nach apt update

Daemon hat 2 loops:
- Einmal am Tag über liste von urls fuer certs
- Einmal pro Stunde über alle clients: Heartbeat ok
