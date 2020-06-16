# Backup

A command line application to backup files from one location to another.

## Build

```
go build -o backup cmd/backup/main.go
mv backup /usr/local/bin
```

## Usage

`backup <source dir> <destination dir>`

Options
```
-f, --fast             | Don't perform hashsum checks on files of the same size (assume their contents are equal by the file size)
-m, --mirror           | Make the destination directory a mirror of the source directory (Removes any files in dest that aren't also in source)
-i, --include-symlinks | Also backup any symlinks
-v, --verbose          | Enable debug logging (Warning, lots of logs)
-h, --help             | Print usage
```

## Restoring a backed up directory

Use the `-m, --mirror` flag to mirror the backup directory to the restore location
`backup -m <backup dir> <dest dir>`