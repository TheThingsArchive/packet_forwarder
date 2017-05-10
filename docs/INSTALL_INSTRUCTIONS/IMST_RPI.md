# Install the TTN Packet Forwarder on a Raspberry Pi with an IMST ic880a board

To follow this manual, you must have a Raspberry Pi with an IMST ic880a board, connected through SPI or USB.

*Note: USB support is experimental, and will be dropped in the future. See [the FTDI documentation](FTDI.md) for more information on the libraries necessary for FTDI support.*

## Download and run

1. Download the [Raspberry Pi + IMST build](https://ttnreleases.blob.core.windows.net/packet_forwarder/master/imst-rpi-pktfwd.zip) of the packet forwarder. The archive will contain both the SPI and USB (FTDI) binaries of the packet forwarder - use the one adapted to your setup.

2. Configure the packet forwarder:

```bash
$ <packet-forwarder-binary> configure
[...]
  INFO New configuration file saved             ConfigFilePath=/root/.pktfwd.yml
```

3. Run the packet forwarder:

```bash
$ <packet-forwarder-binary> start
```

## <a name="build"></a>Build

If want to contribute to the development of the packet forwarder, you might want to build the TTN Packet Forwarder. You will need to use a Linux environment to run the toolchain necessary for the build.

### Getting the toolchain

If you want to build the packet forwarder for a Raspberry Pi, you will need a **Raspberry Pi cross-compiler**. On some Linux distributions, such as Ubuntu, a toolchain is available as a package: `sudo apt install arm-linux-gnueabihf-gcc -y`.

### Building the binary

Make sure you have [installed](https://golang.org/dl/) and [configured](https://golang.org/doc/code.html#GOPATH) your Go environment.

Follow these commands:

```bash
$ make dev-deps
$ make deps
$ GOOS=linux GOARCH=arm GOARM=7 CC=arm-linux-gnueabihf-gcc make build
```

The binary will then be available in the `release/` folder.
