package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"

	"github.com/jpoirier/gousb/usb"
)

const (
	vendorID          = 0x046d // Logitech, Inc.
	productID         = 0xc092 // G102 and G203 Prodigy Gaming Mouse (or G203 Lightsync which G203L likely is)
	packetSize        = 20
	defaultRate       = 10000 // milliseconds
	defaultBrightness = 100   // percentage

	// Mode IDs (these are common for Logitech, verify if you have specific G203L defines)
	modeStatic  byte = 0x01
	modeCycle   byte = 0x02
	modeBreathe byte = 0x03
	// modeWave     byte = 0x04 // Not implemented in this script
	// modeColorMix byte = 0x05 // Not implemented in this script
)

var (
	debug = flag.Int("debug", 0, "libusb debug level (0..3)")
)

func main() {
	flag.Usage = func() {
		fmt.Print(`Logitech G102/G203/G203L Mouse LED control (Revised)

Usage:
  gled solid <color>                         Solid color mode
  gled cycle <rate> <brightness>             Cycle through all colors
  gled breathe <color> <rate> <brightness>   Single color breathing
  gled intro <toggle>                        Enable/disable startup effect (Note: command 0x5B is not from OpenRGB's SetMode)

Arguments:
  color        RRGGBB (RGB hex value, e.g., FF0000 for red)
  rate         100-60000 (Number of milliseconds. Default: 10000ms)
  brightness   1-100 (Percentage. Default: 100%)
  toggle       on|off

Flags:
  gled -debug <0..3> ...                     Debug level for libusb. Default: 0
`)
	}

	flag.Parse()
	mode := flag.Arg(0)
	switch mode {
	case "solid":
		setSolid()
	case "cycle":
		setCycle()
	case "breathe":
		setBreathe()
	case "intro":
		setIntro()
	default:
		flag.Usage()
		log.Fatalf("Unknown mode: %q", mode)
	}
}

// sendRawPayload sends the raw byte payload to the device.
// It now also includes the software control enable packet.
func sendRawPayload(payload []byte) {
	if len(payload) != packetSize {
		log.Fatalf("Error: payload length must be %d bytes, got %d", packetSize, len(payload))
	}

	ctx := usb.NewContext()
	defer ctx.Close()
	ctx.Debug(*debug)

	dev, err := ctx.OpenDeviceWithVidPid(vendorID, productID)
	if err != nil {
		log.Fatalf("Error opening device: %v", err)
	}
	defer dev.Close()

	// Send software control enable packet (from OpenRGB constructor)
	initPacket := make([]byte, packetSize)
	initPacket[0] = 0x11
	initPacket[1] = 0xFF
	initPacket[2] = 0x0E
	initPacket[3] = 0x50 // Enable software control command
	initPacket[4] = 0x01
	initPacket[5] = 0x03
	initPacket[6] = 0x07
	// Remaining bytes are 0x00 by default

	log.Printf("Sending software control enable packet: %x", initPacket)
	_, err = dev.Control(0x21, 0x09, 0x0211, 0x01, initPacket)
	if err != nil {
		// Log as warning as some devices might work without it or if already set
		log.Printf("Warning: Error sending software control enable packet: %v", err)
	}
	// A small delay might be beneficial here if issues occur, e.g., time.Sleep(50 * time.Millisecond)

	log.Printf("Sending command payload: %x", payload)
	n, err := dev.Control(0x21, 0x09, 0x0211, 0x01, payload) // Report ID 0x11, Interface 0x01
	if err != nil {
		log.Fatalf("Error sending control data: %v", err)
	}

	log.Printf("%d bytes transferred for command payload", n)
}

func setIntro() {
	toggleStr := flag.Arg(1)
	if toggleStr == "" {
		flag.Usage()
		log.Fatal("Missing toggle argument for intro mode")
	}
	toggleVal := parseToggleByte(toggleStr)

	payload := make([]byte, packetSize)
	payload[0] = 0x11
	payload[1] = 0xFF
	payload[2] = 0x0E
	payload[3] = 0x5B // Specific command for "intro" effect
	payload[4] = 0x00
	payload[5] = 0x01
	payload[6] = toggleVal
	// payload[7-19] are 0x00 by default

	log.Println("Setting intro effect (Note: command 0x5B is not from OpenRGB's SetMode logic)")
	sendRawPayload(payload)
}

func setSolid() {
	colorStr := flag.Arg(1)
	if colorStr == "" {
		flag.Usage()
		log.Fatal("Missing color argument for solid mode")
	}
	r, g, b := parseHexColorToRGB(colorStr)

	payload := make([]byte, packetSize)
	payload[0] = 0x11
	payload[1] = 0xFF
	payload[2] = 0x0E
	payload[3] = 0x10 // SetMode command
	payload[4] = 0x00
	payload[5] = modeStatic
	payload[6] = r
	payload[7] = g
	payload[8] = b
	payload[9] = 0x02 // Mode specific data for static (from OpenRGB C++ code)
	// payload[10-15] are 0x00
	payload[16] = 0x01 // End byte
	// payload[17-19] are 0x00

	sendRawPayload(payload)
}

func setCycle() {
	rateStr := flag.Arg(1)
	brightnessStr := flag.Arg(2)
	if rateStr == "" || brightnessStr == "" {
		flag.Usage()
		log.Fatal("Missing rate or brightness argument for cycle mode")
	}

	rate := parseRate(rateStr)
	brightness := parseBrightness(brightnessStr) // 1-100

	payload := make([]byte, packetSize)
	payload[0] = 0x11
	payload[1] = 0xFF
	payload[2] = 0x0E
	payload[3] = 0x10 // SetMode command
	payload[4] = 0x00
	payload[5] = modeCycle
	payload[6] = 0x00 // Red for cycle (typically 0)
	payload[7] = 0x00 // Green for cycle
	payload[8] = 0x00 // Blue for cycle

	// Speed (rate)
	payload[11] = byte(rate >> 8)   // Speed Hi
	payload[12] = byte(rate & 0xFF) // Speed Lo

	// Brightness
	deviceBrightness := brightness * 5
	// if deviceBrightness == 0 { deviceBrightness = 1; } // Not needed for brightness 1-100 input
	payload[13] = byte(deviceBrightness) // Implicitly (value % 256)

	payload[16] = 0x01 // End byte

	sendRawPayload(payload)
}

func setBreathe() {
	colorStr := flag.Arg(1)
	rateStr := flag.Arg(2)
	brightnessStr := flag.Arg(3)
	if colorStr == "" || rateStr == "" || brightnessStr == "" {
		flag.Usage()
		log.Fatal("Missing color, rate, or brightness argument for breathe mode")
	}

	r, g, b := parseHexColorToRGB(colorStr)
	rate := parseRate(rateStr)
	brightness := parseBrightness(brightnessStr) // 1-100

	payload := make([]byte, packetSize)
	payload[0] = 0x11
	payload[1] = 0xFF
	payload[2] = 0x0E
	payload[3] = 0x10 // SetMode command
	payload[4] = 0x00
	payload[5] = modeBreathe
	payload[6] = r
	payload[7] = g
	payload[8] = b

	// Speed (rate)
	payload[9] = byte(rate >> 8)    // Speed Hi
	payload[10] = byte(rate & 0xFF) // Speed Lo

	// Brightness
	deviceBrightness := brightness * 5
	payload[12] = byte(deviceBrightness) // Implicitly (value % 256)

	payload[16] = 0x01 // End byte

	sendRawPayload(payload)
}

func parseToggleByte(toggleArg string) byte {
	switch strings.ToLower(toggleArg) {
	case "on":
		return 0x01
	case "off":
		return 0x02
	default:
		flag.Usage()
		log.Fatalf("Error parsing toggle argument: %q. Use 'on' or 'off'.", toggleArg)
		return 0 // Should not reach here
	}
}

func parseHexColorToRGB(colorArg string) (r, g, b byte) {
	if colorArg == "" {
		flag.Usage()
		log.Fatal("No color argument found")
	}
	tempColorArg := colorArg
	if !strings.HasPrefix(tempColorArg, "#") {
		tempColorArg = "#" + tempColorArg
	}

	c, err := hexToColor(tempColorArg)
	if err != nil {
		flag.Usage()
		log.Fatalf("Error parsing color argument %q: %v", colorArg, err)
	}
	return c.R, c.G, c.B
}

// hexToColor converts an #RRGGBB or #RGB string to a color.RGBA struct
func hexToColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7: // #RRGGBB
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4: // #RGB
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits: #ABC -> #AABBCC
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid hex color format: %q", s)
	}
	return
}

func parseRate(rateArg string) int {
	var rate int
	if rateArg == "" {
		rate = defaultRate
	} else {
		var err error
		rate, err = strconv.Atoi(rateArg)
		if err != nil {
			flag.Usage()
			log.Fatalf("Error parsing rate argument %q: %v", rateArg, err)
		}
		if rate < 100 || rate > 60000 { // Limits from OpenRGB C++ comments/typical usage
			flag.Usage()
			log.Fatalf("Rate argument %q is out of range (100-60000)", rateArg)
		}
	}
	return rate
}

func parseBrightness(brightnessArg string) int {
	var brightness int
	if brightnessArg == "" {
		brightness = defaultBrightness
	} else {
		var err error
		brightness, err = strconv.Atoi(brightnessArg)
		if err != nil {
			flag.Usage()
			log.Fatalf("Error parsing brightness argument %q: %v", brightnessArg, err)
		}
		if brightness < 1 || brightness > 100 { // Percentage
			flag.Usage()
			log.Fatalf("Brightness argument %q is out of range (1-100)", brightnessArg)
		}
	}
	return brightness
}
