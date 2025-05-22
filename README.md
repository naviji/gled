Logitech G102, G103 and G203 LIGHTSYNC LED control
================================================

[![Go Report Card](https://goreportcard.com/badge/github.com/naviji/gled)](https://goreportcard.com/report/github.com/naviji/gled) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) `gled` is a command-line tool written in Go to control the RGB LED lighting on Logitech G102, G203, and G203L (Prodigy/Lightsync) gaming mice on Linux and macOS.

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
    git clone [https://github.com/naviji/gled.git](https://github.com/naviji/gled.git)
    cd gled
    ```

2.  **Build the executable:**

      * **Attempt standard build first:**
        Go often uses `pkg-config` to find `libusb`. If `libusb` is installed correctly in a standard location or `pkg-config` is set up for it, this might just work:

        ```bash
        go build -o gled gled.go
        ```

      * **If the standard build fails (e.g., `libusb.h` not found):**
        You might need to tell CGo where to find `libusb` header and library files, especially if it's installed in a custom location (like Homebrew on macOS sometimes requires for command-line builds).

        **On macOS (using Homebrew `libusb`):**
        The paths exported in your example are a reliable way:

        ```bash
        export LIBUSB_PREFIX=$(brew --prefix libusb)
        export CGO_CFLAGS="-I${LIBUSB_PREFIX}/include"
        export CGO_LDFLAGS="-L${LIBUSB_PREFIX}/lib"
        go build -o gled gled.go
        ```

        You can add these exports to your shell's configuration file (e.g., `~/.zshrc` or `~/.bash_profile`) if you build frequently.

        **On Linux (if `libusb` is in a non-standard path):**
        Replace `/path/to/libusb` with the actual prefix where `libusb` is installed (containing `include` and `lib` directories).

        ```bash
        export CGO_CFLAGS="-I/path/to/libusb/include"
        export CGO_LDFLAGS="-L/path/to/libusb/lib"
        go build -o gled gled.go
        ```

        However, on most Linux distributions, installing the `libusb-1.0-0-dev` (or equivalent) package should set it up for `pkg-config`, making these exports unnecessary.

3.  **Place the `gled` executable in your PATH (optional):**

    ```bash
    sudo mv gled /usr/local/bin/
    ```

## Permissions

Controlling USB devices typically requires special permissions.

  * **Linux:** You'll likely need to run `gled` with `sudo`:

    ```bash
    sudo ./gled solid ff0000
    ```

    Alternatively, to run `gled` as a non-root user, you can set up `udev` rules. Create a file named `/etc/udev/rules.d/99-logitech-gled.rules` with the following content:

    ```
    SUBSYSTEM=="usb", ATTR{idVendor}=="046d", ATTR{idProduct}=="c092", MODE="0666"
    ```

    Then reload the udev rules:

    ```bash
    sudo udevadm control --reload-rules
    sudo udevadm trigger
    ```

    You might need to unplug and replug your mouse for the rules to take effect.

  * **macOS:** Usually, `sudo` is not required if `libusb` is installed correctly via Homebrew and you are running the command as the logged-in user.

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

## Troubleshooting

  * **`libusb.h: No such file or directory` during build:**
    Ensure `libusb` development headers are installed (e.g., `libusb-1.0-0-dev` on Debian/Ubuntu, `libusb1-devel` on Fedora, `libusb` via Homebrew on macOS). If they are installed in a non-standard location, use the `CGO_CFLAGS` and `CGO_LDFLAGS` environment variables as described in the "Installation" section.

  * **"Error opening device" or permission denied at runtime:**
    On Linux, try running with `sudo` or set up `udev` rules as described in the "Permissions" section. On macOS, ensure `libusb` was installed correctly.

  * **Command doesn't change LEDs:**

      * Verify your mouse is one of the supported models (VID `046d`, PID `c092`).
      * Try unplugging and replugging the mouse.
      * Use the `-debug 3` flag to get more verbose output from `libusb`, which might indicate communication issues.

## Contributing

Contributions are welcome\! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](https://www.google.com/search?q=LICENSE.txt). \`\`\`
