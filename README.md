# picard

Picard is a Go client for
[Sheepdog](https://sheepdog.github.io/sheepdog/).

## Usage

```go
  // Create a new Picard connection to Sheepdog cluster
  // using the host:port string.
  c, err := picard.NewCluster(hostport)
  if err != nil {
    return err
  }

  // Always make sure to cleanly exit.
  defer c.Disconnect()

  // Open an existing VDI image for reading or writing
  vdi, err := c.OpenVDI("WoolWag")
  if err != nil {
    return err
  }
  defer vdi.Close()

  // Write some bytes onto the VDI.
  if _, err := vdi.Write(someBytes); err != nil {
    return err
  }

  // Read information off of VDI.
  buf := make([]byte, 100)
  if _, err := vdi.Read(buf); err != nil {
    return err
  }

  // Do something with the `buf` here.
```

## Tests

Picard uses the typical way of running Go tests.

```bash
$ go test -tags integration -v
=== RUN TestVDIRun
--- PASS: TestVDIRun (9.99s)
PASS
coverage: 88.6% of statements
ok  github.com/neurodrone/picard 10.002s
```
