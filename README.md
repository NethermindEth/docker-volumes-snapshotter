# Docker Volumes Snapshotter

The snapshotter project was created to fulfill the requirement of saving all the volumes of a Docker container to a tar file, which can be restored at a later time. By properly configuring it, volumes from different containers can be saved in the same tar file.

> The tar file can either already exist or be empty. An empty tar file is not a zero-byte file. To understand this better, refer to the [Initialize `tar` file](#initialize-tar-file) section.

The snapshotter container should run with the same volumes as the container whose data should be saved. One way to achieve this is by running the snapshotter container with the `-volumes-from` flag.

- [Docker Volumes Snapshotter](#docker-volumes-snapshotter)
  - [Build snapshotter image](#build-snapshotter-image)
  - [Backup](#backup)
  - [Restore](#restore)
  - [Configuration file](#configuration-file)
    - [Passing the configuration file](#passing-the-configuration-file)
    - [Configuration format](#configuration-format)
    - [Example](#example)
  - [Backup file](#backup-file)
    - [Initialize `tar` file](#initialize-tar-file)
      - [Using the `backuptar` package](#using-the-backuptar-package)
      - [Using a CLI command](#using-a-cli-command)
    - [Passing the backup `tar` file](#passing-the-backup-tar-file)

![diagram](img/snapshotter-diagram.png)

## Build snapshotter image

```yaml
docker build -t snapshotter:v0.2.0 github.com/NethermindEth/docker-volumes-snapshotter.git#v0.2.0
```

## Backup

To backup volumes from a Docker container use the `backup` command, bind the volumes, configuration file and the `backup.tar` file.

```bash
docker run \
  --rm \
  --volumes-from <container> \
  -v $(pwd)/backup.tar:/backup.tar \
  -v $(pwd)/config.yml:/config.yml \
  eigenlayer-snapshotter:v0.2.0 backup
```

## Restore

To restore volumes of a Docker container use the `restore` command, bind volumes, configuration file and the `backup.tar` file.

```bash
docker run \
  --rm \
  --volumes-from <container> \
  -v $(pwd)/backup.tar:/backup.tar \
  -v $(pwd)/config.yml:/config.yml \
  eigenlayer-snapshotter:v0.2.0 restore
```

## Configuration file

### Passing the configuration file

The snapshotter process requires the configuration file to be located at the `/config.yml` path inside the container. Therefore, the recommended way to pass the configuration is by mounting the following volume:

```text
--volume <path-to-config>:/config.yml
```

Replace `<path-to-config>` with the absolute path to the configuration file on the host machine.

### Configuration format

The snapshotter does not need too many configurations, only two options are necessary:

1. `prefix`: is the prefix path to store the volumes inside the backup tarball file
2. `volumes`: list of volume targets inside the container, should be absolute paths to a directory or a file inside the container

### Example

The following example configures the snapshotter to save volumes under the `prefix/path` inside the `backup.tar` file. Volumes could be directories like `volume1` and `volume3` or files like `volume2.txt`.

```yaml
prefix: "prefix/path"
volumes:
 - /path/to/volume1
 - /path/to/volume2.txt
 - /path/to/volume3
```

## Backup file

### Initialize `tar` file

The backup tar file should be empty or not. However, in any case, it should end with 1024 zero bytes. These zero bytes consist of two empty blocks of 512 bytes each, which define an end-of-archive, as specified in [the spec](https://www.gnu.org/software/tar/manual/html_node/Standard.html). This requirement is necessary to enable the appending of files to existing tar files that contain volumes from other containers, but are located at different paths within the tarball.

The initialization of a backup tar file involves creating a file with a `.tar` extension, with the content being 1024 zero bytes.

#### Using the `backuptar` package

The backup tar file could be initialized using the `InitBackupTar` in the `[backuptar](https://github.com/NethermindEth/docker-volumes-snapshotter/tree/main/pkg/backuptar)` package. For instance:

```go
package main

import (
 "fmt"
 "io"
 "log"
 "os"

 "github.com/NethermindEth/docker-volumes-snapshotter/pkg/backuptar"
)

func main() {
 err := backuptar.InitBackupTar("backup.tar")
 if err != nil {
  log.Fatal(err)
 }

 backupStat, err := os.Stat("backup.tar")
 if err != nil {
  log.Fatal(err)
 }

 fmt.Printf("Backup size: %d\n", backupStat.Size())

 backupF, err := os.Open("backup.tar")
 if err != nil {
  log.Fatal(err)
 }
 defer backupF.Close()

 backupData, err := io.ReadAll(backupF)
 if err != nil {
  log.Fatal(err)
 }

 fmt.Printf("Backup data:\n%v\n", backupData)
}
```

```text
Backup size: 1024
Backup data:
[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
```

#### Using a CLI command

The following command generates a `backup.tar` file with two blocks of 512 zero bytes

```bash
dd if=/dev/zero of=backup.tar bs=512 count=2
```

```text
2+0 records in
2+0 records out
1024 bytes (1.0 kB, 1.0 KiB) copied, 0.00199071 s, 514 kB/s
```

### Passing the backup `tar` file

The snapshotter process requires the backup tar file to be located at the `/backup.tar` path inside the cotainer. Therefore, the proper way to pass the configuration is by mounting the following volume:

```text
--volume <path-to-backup-tar>:/backup.tar
```

Replace `<path-to-backup-tar>` with absolute path to the configuration file on the host machine.
