# Game Streams

[![wakatime](https://wakatime.com/badge/github/tbeachill/game-streams.svg)](https://wakatime.com/badge/github/tbeachill/game-streams)

![Go](https://img.shields.io/badge/-Go-00ADD9?logo=go&logoColor=white&style=flat-square)
![SQLite](https://img.shields.io/badge/-SQLite-003B57?logo=sqlite&logoColor=white&style=flat-square)

## Intro
Game Streams is a Discord bot that keeps track of upcoming streams announcing new games. It can be configured to announce when a stream is about to start for the server's followed platforms.

## Features
- Announces when a stream is about to start to a specified channel and role.
- Users and servers can be blacklisted.
- The database is encrypted and backed up automatically.
- Automatic database maintenance is performed.
- A range of options can be configured in a config.toml file.
- Streams can be batch uploaded as a TOML file.
- Basic analytics about command usage and server membership are collected.

## Commands
- `/streams` displays a list of upcoming streams.
- `/streaminfo` displays all information for a specified stream.
- `/suggest` allows streams to be suggested to be added to the database.
- `/settings` allows announcement settings to be configured.
- `/help` displays help for the bot and each command.
