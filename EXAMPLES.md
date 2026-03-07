# Examples

Detailed command examples for `rollbar-cli`.

## Items

### List items

```bash
# text/TUI output
rollbar-cli items list --status active --environment production

# stable JSON output
rollbar-cli items list --json

# raw Rollbar API JSON
rollbar-cli items list --raw-json

# NDJSON for scripting
rollbar-cli items list --ndjson --limit 20

# page, time, and sort filtering (repeat --level)
rollbar-cli items list --page 2 --pages 3 --level error --level critical --last 24h --sort counter_desc --limit 25

# watch the list during incident triage
rollbar-cli items watch --status active --environment production --interval 30s --count 10
```

### Get one item

```bash
# get a single item by numeric item ID
rollbar-cli items get 275123456
# or
rollbar-cli items get --id 275123456

# get a single item by UUID
rollbar-cli items get 01234567-89ab-cdef-0123-456789abcdef
# or
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef

# get item JSON (stable schema)
rollbar-cli items get --id 275123456 --json

# get item + instance details with payload shaping
rollbar-cli items get --id 275123456 --instances --payload summary --payload-section request

# get item + instances JSON payloads
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --instances --json
```

### Update an item

```bash
# common task verbs
rollbar-cli items resolve --id 275123456 --resolved-in-version aabbcc1
rollbar-cli items mute --id 275123456
rollbar-cli items assign --id 275123456 --assigned-user-id 321
rollbar-cli items snooze --id 275123456 --duration 1h

# update status + resolved version
rollbar-cli items update --id 275123456 --status resolved --resolved-in-version aabbcc1

# update by UUID and set level/title
rollbar-cli items update 01234567-89ab-cdef-0123-456789abcdef --level error --title "Checkout failure"

# clear assignment and set snooze
rollbar-cli items update --id 275123456 --clear-assigned-user --snooze-enabled true --snooze-expiration-seconds 3600

# update item JSON
rollbar-cli items update --id 275123456 --status active --json
```

## Occurrences

```bash
# list occurrences for an item
rollbar-cli occurrences list --item-id 275123456
# or
rollbar-cli occurrences list 275123456

# list occurrences JSON payload
rollbar-cli occurrences list --item-uuid 01234567-89ab-cdef-0123-456789abcdef --json

# list occurrences in NDJSON
rollbar-cli occurrences list --item-id 275123456 --ndjson

# get one occurrence by numeric occurrence ID
rollbar-cli occurrences get --id 501
# or
rollbar-cli occurrences get 501

# get one occurrence by UUID
rollbar-cli occurrences get --uuid 89abcdef-0123-4567-89ab-cdef01234567

# alias spelling also works
rollbar-cli occurences get --uuid 89abcdef-0123-4567-89ab-cdef01234567

# get one occurrence JSON
rollbar-cli occurrences get --uuid 89abcdef-0123-4567-89ab-cdef01234567 --json
```

## Users

```bash
# list account users
rollbar-cli users list

# stable JSON output
rollbar-cli users list --json

# get one user by numeric id
rollbar-cli users get 7

# or stable JSON output for one user
rollbar-cli users get --id 7 --json

# NDJSON for one user
rollbar-cli users get --id 7 --ndjson

# raw Rollbar API JSON
rollbar-cli users list --raw-json

# NDJSON for scripting
rollbar-cli users list --ndjson
```

## Deploys

```bash
# list deploys
rollbar-cli deploys list

# page through deploy history
rollbar-cli deploys list --page 2 --limit 20 --json

# get one deploy by id
rollbar-cli deploys get 12345
# or
rollbar-cli deploys get --id 12345 --json

# create a deploy record
rollbar-cli deploys create \
  --environment production \
  --revision aabbcc1 \
  --status started \
  --comment "Deploy started from CI" \
  --local-username ci-bot

# create a deploy and associate a Rollbar username
rollbar-cli deploys create \
  --environment production \
  --revision aabbcc1 \
  --rollbar-username dave \
  --json

# update a deploy after completion
rollbar-cli deploys update 12345 \
  --status succeeded \
  --json
```

## Environments

```bash
# list all environments across every API page
rollbar-cli environments list

# stable JSON output
rollbar-cli environments list --json

# raw Rollbar API page envelopes
rollbar-cli environments list --raw-json

# NDJSON for scripting
rollbar-cli environments list --ndjson
```

## Shell completion

```bash
# bash, fish, zsh, and powershell are supported
rollbar-cli completion bash
rollbar-cli completion zsh
rollbar-cli completion fish
rollbar-cli completion powershell
```

Examples:

```bash
# bash one-off
source <(rollbar-cli completion bash)

# bash global
rollbar-cli completion bash > rollbar-cli.bash
sudo cp rollbar-cli.bash /etc/bash_completion.d/

# zsh one-off
source <(rollbar-cli completion zsh)

# zsh global
rollbar-cli completion zsh > _rollbar-cli
sudo cp _rollbar-cli /usr/local/share/zsh/site-functions/

# zsh user only
mkdir -p ~/.zsh/completions
rollbar-cli completion zsh > ~/.zsh/completions/_rollbar-cli
```
