```
$ ./euigen help
NAME:
   euigen - A cli application to generate new batches of devEUIs

USAGE:
   euigen [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --discard, -d  Discard last incomplete run if present, instead of resuming (default: false)
   --help, -h     show help (default: false)

```
# Usage
```
$ go build -o euigen main.go  # build
$ ./euigen 100  # generate a batch of 100 DevEUIs 
```