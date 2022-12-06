Telegram Bot, which reads logs from UFW and sends notifications to telegram channel when there are blocked/audited/allowed connections. Run with sudo.

Bot reads the configuration file, parses the path to the file to be followed, parses the Telegram bot API token, parses the channel to which updates are sent, and does other things. After that, the bot uses this information to wait for new changes in the log file and sends messages to the channel in pretty format.

# Usage

## Setup UFW firewall

### Add rules to UFW

### Enable UFW

### Restart server (optional)

## Setup Telegram Bot

### Create Telegram Bot

### Copy Bot API Token

### Create channel with logs

### Check channel's ID

Send a message to the new channel and copy link of the message. You'll find a channel ID in the link (between two slashes). Add `-100` before the ID and note it for the future.

## Build project

Just clone this project to your server and run `go build`. Otherwise you can use prebuilt binary from releases (if it persists).

## Setup config

Add the Bot API token and channel's ID to the config. Use `config-example.yml` template from the root of the repository.

Don't forget to rename a config file.

Add vital ports, which are important ports you want to distinguish in the channel with updates.

## Run the program

1. Run the executable with sudo
2. Use `nohup` or something like that to keep program running even if you're exiting from the server
3. Check if everything is working