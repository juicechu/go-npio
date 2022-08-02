go-npio
=======

Native GPIO-Gophers for your NanoPi!


go-npio is a Go library for accessing [GPIO](https://wiki.friendlyelec.com/wiki/index.php/NanoPi_M1_Plus)-pins
on the [NanoPi M1 Plus].

## Releases ##
- 0.0.1 - Support NanoPi M1 Plus.

#Todo

* [] PWM
* [] Clock
* [] Support NanoPi NEO and other NanoPi

## Usage ##

```go
import "github.com/juice/go-npio"
```

Open memory range for GPIO access in /dev/mem

```go
err := npio.Open()
```

Initialize a pin, run basic operations.
Pin refers to the wPi pin, not the physical pin on the NanoPi header. Pin 10 here is exposed on the pin header as physical pin 24.

```go
pin := npio.Pin(10)

pin.Output()       // Output mode
pin.High()         // Set pin High
pin.Low()          // Set pin Low
pin.Toggle()       // Toggle pin (Low -> High -> Low)

pin.Input()        // Input mode
res := pin.Read()  // Read state from pin (High / Low)

pin.Mode(npio.Output)   // Alternative syntax
pin.Write(npio.High)    // Alternative syntax
```

Pull up/down/off can be set using:

```go
pin.PullUp()
pin.PullDown()
pin.PullOff()

pin.Pull(rpio.PullUp)
```

Unmap memory when done

```go
npio.Close()
```

## Other ##

Currently, it supports basic functionality such as:
- Pin Direction (Input / Output)
- Write (High / Low)
- Read (High / Low)

It works by memory-mapping the NanoPi M1 Plus gpio range, and therefore require root/administrative-rights to run.
