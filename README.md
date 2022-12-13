# baton-AWS-discord
[baton-aws](https://github.com/ConductorOne/baton-aws) is a connector for AWS built using the Baton SDK. It communicates with the AWS API to sync data about which groups and users have access to accounts, groups, and roles within an AWS org.

**baton-AWS-discord** allows for usage with a discord bot. Please follow bot setup instructions on discord via [discord bot setup](https://discord.com/developers/applications). Make sure to enable **MESSAGE CONTENT INTENT**.

## Installation
1. Install [baton-aws](https://github.com/ConductorOne/baton-aws)
2. Clone this repo 
3. Install go dependencies or run `go get` in the root directory
4. Run `go build` in the root directory
5. Run via `go run main.go` or via `./main` in the terminal

**Note: make sure to create .env file in root dir and set your discord bot token as DISCORD_TOKEN = "<YOUR_TOKEN_HERE>"**
