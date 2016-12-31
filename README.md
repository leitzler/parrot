# parrot
This is a Slack bot that adds group notifications using private messages. A quick hack for demo purpose only.

## Running
Copy config_template.json to config.json and set the `"admin"` field to your Slack user ID (e.g. `"U01ABCDEF9"`).

Get yourself a Slack bot API token, and set the environment variable `PARROT_SLACK_TOKEN`:
```
$ PARROT_SLACK_TOKEN=abcd-123456789012-123456789012345678901234 go run parrot.go
```
## Usage
The bot listens to messages in all channels it is invited to. If a message starts with a word with `@` as first letter, parrot checks if that trigger word exists and has any receivers. If so, it forwards the entire message to all of them as private message.

## Commands
Commands are sent to parrot as private messages.

#### Available to anyone
- `list` - lists all trigger words & receivers

#### Available to admin
- `set` - set a trigger word, usage example `set foobar @james @jane` creates the `foobar` trigger word.
- `del` - delete a trigger word, usage example `del foobar`.
- `save` - save the config to file
- `debug` - enable debug printouts
