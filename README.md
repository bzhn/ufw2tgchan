Telegram Bot which reads configs from ufw and sends notifications there are blocked/audited/allowed connections. Use with sudo.

Bot reads config file, parses path to file to follow, telegram bot API token, channel to send updates to and something else. After that bot uses this information to wait for new changes in the log file and sends messages to the channel in pretty format.

# Example

Here is the example of messages sent to the channel.

![image](https://user-images.githubusercontent.com/89320434/205466971-4120b8c1-6df1-4ccd-aad3-9c055d6974fa.png)

Note the vital port that was distinguished with !! emoji. I setted this up in the settings, when had added the Redis's port (6379) to vital ports in `config.yml`.


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
2. Use `nohup` or something like that to keep program running even if you're exiting from the server.
3. Check if everything is working.

