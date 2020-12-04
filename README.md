# dndbeyong-hp

Get your characters hp from dndbeyond!

## Download

Download from release page.

THIS PROGRAM REQUIRE CHROME INSTALLED.

If you are running linux or WSL:

```bash
wget http://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo dpkg -i google-chrome*.deb
```

## How To Use

### run directly

Run executable directly and it will ask you for your character ID. It will take 5 sec to populate the table.

### run CLI

CLI is also possible.

```bash
./dndbeyong-hp --help

NAME:
   dndbeyond-hp - Get your characters hp from dndbeyond! Please set all characters to public.

USAGE:
   {Character ID} - Your dndbeyond character id

GLOBAL OPTIONS:
   --interval value, -i value  Set refresh interval. Not recommend to set lower as dndbeyond has DDOS protection. Example: https://godoc.org/github.com/robfig/cron (default: "@every 1m")
   --help, -h                  show help (default: false)
```

## How to build

### linux

```bash
env GOOS=linux GOARCH=amd64 go build
```

### macos

```bash
env GOOS=darwin GOARCH=amd64 go build
```

### windows

```bash
env GOOS=windows GOARCH=amd64 go build
```

## How?

As you know, dndbeyond doesn't provide any API whatsoever. This program basically use headless chrome render the character sheet, and use queryselector to get information.

## I want to help!

If you are interested in this project, please leave comment in the issue page.

## TODO

- add `-o` flag to cli, so that cli can output json, or text. Programs like jq or grep can benefit from this.
- add more information that we can scrap from character sheet.
