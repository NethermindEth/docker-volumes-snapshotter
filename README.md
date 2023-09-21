# Docker Volumes Snapshotter

This repository is useful for create a snapshot of a list of volumes into a
tar file.

- [Docker Volumes Snapshotter](#docker-volumes-snapshotter)
  - [Build snapshotter image](#build-snapshotter-image)
  - [Running the snapshotter](#running-the-snapshotter)
    - [Example](#example)

![diagram](img/snapshotter-diagram.png)

## Build snapshotter image

To build the image using this public repository as context you can use the
following command:

```bash
docker build -t docker-volumes-snapshotter github.com/NethermindEth/docker-volumes-snapshotter.git
```

If you prefer, you can use the Dockerfile from this repository directly cloning
the repository first:

```bash
git clone github.com/NethermindEth/docker-volumes-snapshotter.git
cd docker-volumes-snapshotter
docker build -t docker-volumes-snapshotter .
```

## Running the snapshotter

To run the snapshotter a configuration file is required. This file is a YAML file with the following structure:

```yaml
prefix: "prefix/path/to/store/the/backup"
out: "/path/to/backup.tar"
volumes:
  - "/path/to/volume1"
  - "/path/to/volume2"
```

The snapshotter tries to append the volumes to the backup tar file. If the tar file does not exists, or is empty, an error is raised. To initialize a proper empty tar file run the following command:

```bash
tar -cvf backup.tar --files-from /dev/null
```

You can run the snapshotter different times targeting the same output tar file using different prefix paths. This is useful to create backups of different containers in the same tar file.

### Example

Add the following configuration file to the root of this repository as `config.yml`:

```yaml
prefix: volumes/mock-avs-second
out: /backup.tar
volumes:
  - /cli
  - /README.md
```

Run following command in the root of this repository:

```bash
docker run --rm \
    -v $(pwd)/backup.tar:/backup.tar \       # Bind the backup tar
    -v $(pwd)/config.yml:/snapshotter.yml \  # Bind the configuration file
    -v $(pwd)/cli:/cli \                     # Bind a volume to backup
    -v $(pwd)/README.md:/README.md \         # Bind a file to backup
    docker-volumes-snapshotter
```

Checking the content of the tar file:

```bash
tar -tvf backup.tar
```

Output:

```bash
drwxr-xr-x  0 root   root        0 Sep 21 14:53 volumes/mock-avs-second/1821030e0e15cc096c6c3a13c936baad1b8a4a98f0cbaf2b25cd4642203093fd
-rw-r--r--  0 root   root      653 Sep 22 11:46 volumes/mock-avs-second/1821030e0e15cc096c6c3a13c936baad1b8a4a98f0cbaf2b25cd4642203093fd/cli.go
-rw-r--r--  0 root   root     1929 Sep 22 14:52 volumes/mock-avs-second/95e0a42d9d6b5d82bdb3752f4d31f3fe7d0150c6b512bc094985b1a2b24b192b
-rw-------  0 root   root      203 Sep 22 14:52 volumes/mock-avs-second/volumes-data.yml
```

Notice the presence of the `volumes-data.yml` file. This file contains metadata about the list of volumes in the prefix path. This file is used by the restore process to restore the volumes in the same path. The content of this file is:

```yaml
- id: 1821030e0e15cc096c6c3a13c936baad1b8a4a98f0cbaf2b25cd4642203093fd
  type: dir
  target: /cli
- id: 95e0a42d9d6b5d82bdb3752f4d31f3fe7d0150c6b512bc094985b1a2b24b192b
  type: file
  target: /README.md
```
