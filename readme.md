```
NAME:
   euigen - A cli application to generate new batches of devEUIs

USAGE:
   euigen [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --resume, -r  Resume last incomplete run if present (default: true)
   --help, -h    show help (default: false)
```
# Usage
```
$ go build -o euigen main.go  # build
$ ./euigen 100  # generate a batch of 100 DevEUIs 
```