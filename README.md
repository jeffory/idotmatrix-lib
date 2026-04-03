# idotmatrix-lib

A Go library and CLI for controlling [iDotMatrix](https://www.idotmatrix.com/) pixel displays over Bluetooth Low Energy.

Supports displaying text, images, GIFs, drawing pixels, syncing time, and generating pixel art with AI — all from the command line.

## Platform Support

| Platform | Bluetooth Stack | CGo Required |
|----------|----------------|--------------|
| Linux    | BlueZ via D-Bus | No |
| macOS    | CoreBluetooth   | Yes |
| Windows  | WinRT           | No |

## Installation

```sh
go install github.com/jeffory/idotmatrix-lib/cmd/idotmatrix@latest
```

Or build from source:

```sh
git clone https://github.com/jeffory/idotmatrix-lib.git
cd idotmatrix-lib
go build -o idotmatrix ./cmd/idotmatrix
```

### Prerequisites

- **Linux**: BlueZ (`sudo apt install bluez` / `sudo dnf install bluez`)
- **macOS**: Xcode command line tools (`xcode-select --install`)
- **Windows**: No additional dependencies

## Quick Start

```sh
# Find nearby devices
idotmatrix discover

# Save a device for future use
idotmatrix discover --save

# Display text
idotmatrix text "Hello!" --color ff0000

# Show an image
idotmatrix image photo.png

# Play a GIF
idotmatrix gif animation.gif
```

## Usage

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-a, --address` | Device MAC address | Config/env |
| `-s, --size` | Display size: `16`, `32`, `64` | `32` |
| `--timeout` | Connection timeout (seconds) | `10` |
| `--contrast` | Image contrast (-100 to 100) | `0` |
| `--saturation` | Image saturation (-100 to 100) | `0` |

The device address can also be set via the `IDOTMATRIX_ADDRESS` environment variable or saved to config with `discover --save`.

### Commands

#### `discover`

Scan for nearby iDotMatrix devices.

```sh
idotmatrix discover
idotmatrix discover --save   # Save first device to config
```

#### `text [message]`

Display text on the device.

```sh
idotmatrix text "Hello World" --mode scroll-left --speed 80 --color 00ff00
idotmatrix text "Smooth" --smooth --speed 50
```

| Flag | Description | Default |
|------|-------------|---------|
| `--mode` | Animation mode (see below) | `fixed` |
| `--speed` | Speed (0-255, or ms/frame with `--smooth`) | `80` |
| `--color` | Hex color | `ffffff` |
| `--smooth` | Smooth pixel-by-pixel scrolling | `false` |

**Text modes**: `fixed`, `scroll-left`, `scroll-right`, `scroll-up`, `scroll-down`, `strobe`, `fade`, `falling`, `laser`

#### `image [file]`

Display a static image. Supports PNG, JPG, and other common formats. Automatically resized to fit the display.

```sh
idotmatrix image photo.png
idotmatrix image logo.jpg --contrast 20 --saturation -10
```

#### `gif [file]`

Upload and play a GIF animation.

```sh
idotmatrix gif animation.gif
```

#### `pixel`

Draw a single pixel on the display.

```sh
idotmatrix pixel --x 10 --y 15 --color ff0000
```

#### `power [on|off]`

Turn the display on or off.

```sh
idotmatrix power off
idotmatrix power on
```

#### `time`

Sync the device clock with your system time.

```sh
idotmatrix time
```

#### `reset`

Reset the device.

```sh
idotmatrix reset
```

#### `generate [description]`

Generate pixel art using the [Pixellab](https://pixellab.ai/) API and display it.

```sh
export PIXELLAB_API_KEY=your-key
idotmatrix generate "a tiny red dragon"
idotmatrix generate "blue sword" --no-background --outline soft --save sword.png
```

| Flag | Description | Default |
|------|-------------|---------|
| `--api-key` | Pixellab API key (or `PIXELLAB_API_KEY` env) | |
| `--negative` | What to avoid in generation | |
| `--guidance` | Text guidance scale (1.0-20.0) | `8.0` |
| `--no-background` | Transparent background | `false` |
| `--outline` | Outline style | |
| `--shading` | Shading style | |
| `--detail` | Detail level | |
| `--seed` | Random seed (-1 = random) | `-1` |
| `--save` | Save generated PNG to file | |

#### `animate [description] [action]`

Generate animated pixel art using Pixellab and display it.

```sh
idotmatrix animate "a knight" "walking" --frames 8 --direction east
idotmatrix animate "a cat" "sleeping" --view side --save cat.gif
```

| Flag | Description | Default |
|------|-------------|---------|
| `--frames` | Number of animation frames | `4` |
| `--delay` | Frame delay in centiseconds | `10` |
| `--view` | Camera view: `side`, `low top-down`, `high top-down` | `side` |
| `--direction` | Facing direction (e.g., `south`, `east`, `north-west`) | `south` |
| `--guidance` | Guidance scale (1.0-20.0) | `4.0` |
| `--palette` | Forced color palette (comma-separated hex) | |
| `--save` | Save generated GIF to file | |

## Configuration

Configuration is loaded from (highest priority first):

1. Command-line flags
2. Environment variables (`IDOTMATRIX_ADDRESS`, `PIXELLAB_API_KEY`)
3. Config file (`~/.config/idotmatrix/config.yaml`)

Example config:

```yaml
device:
  address: "AA:BB:CC:DD:EE:FF"
image:
  contrast: 0
  saturation: 0
```

## Using as a Library

The project is structured as a set of Go packages that can be used independently:

```go
package main

import (
    "github.com/jeffory/idotmatrix-lib/command"
    "github.com/jeffory/idotmatrix-lib/device"
    "github.com/jeffory/idotmatrix-lib/protocol"
    "github.com/jeffory/idotmatrix-lib/transport"
)

func main() {
    // Connect to a device
    conn, err := transport.Connect("AA:BB:CC:DD:EE:FF", transport.DefaultOptions())
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Create a device handle
    dev := device.New(conn, protocol.Display32x32)

    // Send a command
    cmd := command.NewText("Hello!",
        command.WithTextMode(protocol.TextScrollLeft),
        command.WithSpeed(80),
        command.WithTextColor(protocol.Color{R: 0, G: 255, B: 0}),
    )
    dev.Send(cmd)
}
```

### Display Sizes

- `protocol.Display16x16` — 16x16 pixels
- `protocol.Display32x32` — 32x32 pixels
- `protocol.Display64x64` — 64x64 pixels

### Packages

| Package | Purpose |
|---------|---------|
| `transport` | BLE connection and data transfer |
| `protocol` | Low-level packet encoding (CRC, headers) |
| `command` | High-level command builders (text, image, gif, etc.) |
| `device` | Device abstraction tying commands to transport |
| `font` | TrueType and bitmap font rendering |
| `pixellab` | Pixellab AI API client |

## Examples

Working examples are in the `examples/` directory. Set `IDOTMATRIX_ADDRESS` before running:

```sh
export IDOTMATRIX_ADDRESS="AA:BB:CC:DD:EE:FF"
go run ./examples/discover
go run ./examples/text
go run ./examples/image path/to/image.png
go run ./examples/pixel
```

## License

See [LICENSE](LICENSE) for details.
