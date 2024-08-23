# Awake

A toolkit.

## Quick Start

Download the corresponding executable file for your system from the releases

```bash
# linux/amd64
curl -L -o awake  https://github.com/tianluanchen/awake/releases/download/bin/awake_linux_amd64 && chmod +x awake
```

## Usage

```bash
$ awake --help
A toolkit

Usage:
  awake [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  fetch       Concurrent download of web content to local
  help        Help about any command
  install     Install to GOPATH BIN
  scan        Port scanning
  serve       Start static files server
  tcping      Tcping
  udping      Udping
  unzip       Unarchive zip
  version     Print the version of awake
  zip         Archive files with zip

Flags:
  -h, --help           help for awake
      --level string   log level, DEBUG INFO WARN ERROR FATAL (default "INFO")

Use "awake [command] --help" for more information about a command.
```

## License

[GPL-3.0](./LICENSE) © Ayouth