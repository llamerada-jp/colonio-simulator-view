# simulator-view

Build

```
$ sudo apt install libgl1-mesa-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
$ go build .
```

Help

```
Usage:
  simulator-view [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  plane       View data for plane
  sphere      View data for sphere

Flags:
  -c, --collection string   collection name of mongoDB to get source data (default "logs")
  -d, --database string     database name of mongoDB to get source data (default "simulation")
  -l, --detail-leval uint   Whether to draw detailed information
  -f, --follow              Specify if the logs should be streamed
  -h, --help                help for simulator-view
  -i, --image-name string   Image path and name pattern like hoge/foo@.png (@ will be replace by index like 001, 002...)
  -t, --tail                Output start with tail 10 seconds of the source data
  -u, --uri string          URI of mongoDB to get source data (default "mongodb://localhost:27017")

Use "simulator-view [command] --help" for more information about a command.
```