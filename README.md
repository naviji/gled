Logitech G102, G103 and G203 LIGHTSYNC LED control
================================================

`gled` is a command-line tool written in Go to control the RGB LED lighting on Logitech G102, G203, and G203L (Prodigy/Lightsync) gaming mice on Linux and macOS.

## Features

* Set LEDs to a **solid color**.
* Enable a **color cycle** effect with adjustable speed and brightness.
* Activate a **breathing effect** with a specific color, adjustable speed, and brightness.
* Toggle the **startup intro effect** (on/off).
* Debug output from `libusb`.

## Supported Devices

* Logitech G102, G103 and G203 Lightsync Gaming Mouse

*(Device Product ID: `0xc092`, Vendor ID: `0x046d`)*

## Prerequisites

* **Go:** Version 1.18 or newer is recommended.
* **libusb:** Development library version 1.0 or later. `gousb` (the Go library used by `gled`) requires `libusb` to communicate with USB devices.

## Installation

### 1. Install Prerequisites

#### libusb

* **macOS (using Homebrew):**
    ```bash
    brew install libusb
    ```

### 2. Install GLED
1.  **Clone the repository (if you haven't already):**

    ```bash
    git clone https://github.com/naviji/gled.git
    cd gled
    ```

2.  **Build the executable:**
        ```bash
        go build -o gled gled.go
        ```

## Usage

```bash
gled [global flags] <mode> [mode arguments...]
```

### Modes & Examples

#### Solid Color Mode

Sets the mouse LED to a single static color.

```bash
gled solid <color>
```

Example:

```bash
./gled solid FF0000  # Sets LED to red
./gled solid 00FF00  # Sets LED to green
./gled solid 000000  # Turns LED off (black)
```

#### Cycle Through All Colors

Activates a rainbow color cycle effect.

```bash
gled cycle <rate> <brightness>
```

Example:

```bash
./gled cycle 10000 75  # Cycle at default speed (10000ms), 75% brightness
./gled cycle 5000 100 # Faster cycle, full brightness
```

#### Single Color Breathing

Activates a breathing (pulsing) effect with a chosen color.

```bash
gled breathe <color> <rate> <brightness>
```

Example:

```bash
./gled breathe 0000FF 1500 50 # Breathe blue, 1.5s cycle, 50% brightness
```

#### Enable/Disable Startup Effect

Toggles the mouse's built-in startup lighting effect.
*(Note: This command uses a different USB control message (0x5B) than the other lighting modes, not directly derived from the OpenRGB SetMode function used for other effects.)*

```bash
gled intro <toggle>
```

Example:

```bash
./gled intro off # Disable startup effect
./gled intro on  # Enable startup effect
```

### Arguments

  * **`<color>`:** RGB hex value (e.g., `FF0000` for red, `00FF00` for green).
    Can be 6-digit (RRGGBB) or 3-digit (RGB, e.g., `F00` for red). A `#` prefix is optional.
  * **`<rate>`:** Speed of the effect in milliseconds (Range: 100-60000). Default: `10000ms`.
  * **`<brightness>`:** Brightness percentage (Range: 1-100). Default: `100%`.
  * **`<toggle>`:** `on` or `off`.

### Global Flags

  * **`-debug <level>`:** Set the libusb debug level (0-3). Default: `0` (no debug output).
    Example:
    ```bash
    ./gled -debug 3 solid FF0000
    ```

## Contributing

Contributions are welcome\! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](https://www.google.com/search?q=LICENSE.txt). \`\`\`
