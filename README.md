# xmrig-mo-docker
A tiny docker container for quickly getting up and running with the MoneroOcean fork of xmrig.

![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/thelolagemann/xmrig-mo?style=flat-square)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/thelolagemann/docker-xmrig-mo/Build%20and%20publish%20docker%20image?style=flat-square)
![Xmrig Version](https://img.shields.io/badge/xmrig-v6.15.3-orange?style=flat-square)
### Table of Contents
* [Quick Start](#quick-start)
* [Usage](#usage)
  * [Environment Variables](#environment-variables)
  * [Data Volumes](#data-volumes)
  * [Ports](#ports)
* [Docker Compose](#docker-compose)
* [Notes](#notes)
  * [User/Group IDs](#usergroups-ids)
  * [Configuration](#configuration)
  * [MSR](#msr)
* [Building](#building)

## Quick start
**NOTE**: The command provided is an example and should be adjusted for your needs. 

Launch the miner with the following command:

```bash
docker run -d \
  --name="xmrig-mo" \
  -p "127.0.0.1:3001:3001" \
  -v /cfg/xmrig-mo:/cfg \
  -e WALLET_ADDRESS="88yUzYzB9wrR2r2o1TzXxDMENr6Kbadr3caqKTBUNFZ3dWVt6sJcpWBAwMwNRtEi7nHcBcqzmExNfdNK7ughaCeUFuXXpPp" \
  --restart=always \
  thelolagemann/xmrig-mo:latest
```
Where:
* `/cfg/xmrig-mo`: The directory that will contain any configuration files.
* `WALLET_ADDRESS`: The XMR wallet address to mine to.

Upon first running the container, the miner will perform a short benchmark that should take just a few minutes. These
results are saved in the config.json file under the mounted `/cfg` directory. Once these benchmarks are complete, you
will be able to access the xmrig-workers GUI at `http://host-ip:3001`. 

## Usage
```bash
docker run [-d] \
  --name=xmrig-mo \
  [ -e <VARIABLE_NAME>=<VALUE>]... \
  [-v <HOST_DIR>:<CONTAINER_DIR>[:PERMISSIONS] ]... \
  [-p <HOST_PORT>:<CONTAINER_PORT> ]... \
  thelolagemann/xmrig-mo
```

| **Parameter** |  **Description** |
| --- | --- |
| `-d` | Run the container in the background. If not set, the container runs in the foreground. |
| `-e` | Pass an environment variable to the container. See the [Environment Variables](#env) section for more details. |
| `-v` | Set a volume mapping (allows to share a folder between the host and the container). See the [Data Volumes](#volumes) section for more details |
| `-p` | Set a network port mapping (exposes an internal container port to the host). See the [Ports](#ports) section for more details |

### Environment Variables

| **Variable** | **Description** | **Default** |
| --- | --- | --- |
| `PUID` | User ID of the application. See [User/Group IDs](#usergroups-ids) to better understand when and why this should be set. | `1000` |
| `PGID` | Group ID of the application. See [User/Group IDs](#usergroups-ids) to better understand when and why this should be set. | `1000` |
| `RIG_NAME` | Name used to identify the mining rig. | Randomly generated |
| `API_TOKEN` | API token used to access the xmrig API. | Randomly generated |
| `WALLET_ADDRESS` | The xmr wallet to payout to. | (unset) |
| `XMRIG_API_ENABLED` | Enable the xmrig API. | `true` |
| `XMRIG_WORKERS_ENABLED` | Enable xmrig-workers<sup>[1](#envFt1)</sup> | `true` |
| `XMRIG_WORKERS_AUTOCONFIGURE` | Automatically inject the xmrig api configuration into the xmrig-workers GUI.<sup>[2](#envFt2)</sup> | `true` |
| `BENCHMARK` | Enable benchmarks. By default the benchmarks will only be performed on the initial run. Useful when deploying to environments with dynamically allocated resources. | (unset) |

<sup><a name="envFt1">1</a>: *Enabling xmrig-workers automatically enables the xmrig API*

<sup><a name="envFt2">2</a>: *This will automatically inject the API access token which may pose a security risk if you
have your container exposed to the internet.*

### Data Volumes
The following table describes data volumes used by the container. The mappings are set via the `-v` parameter.

| **Container Path** | **Permissions** | **Description** |
| --- | --- | --- |
| `/cfg` | rw | This is where the miner stores its [configuration](#configuration). |

### Ports
A list of ports used by the container. They can be mapped to the host via the `-p` parameter (one per port mapping).

| Port | Required | Description |
| --- | --- | --- |
| `3000` | No | Port used to query the xmrig API |
| `3001` | No | Port to access xmrig-workers web UI <sup>[1](#xmrigWorkerFootnote)</sup> |

<sup><a name="xmrigWorkerFootnote">1</a>: *Enabling xmrig-workers automatically enables the xmrig API, which may pose 
issues if port `3000` isn't exposed from the container. In order to overcome this, any requests made to 
http://localhost:3001 with the Authorization header present are proxied internally.*</sup>

## Docker Compose

Here is an example of a `docker-compose.yml` file that can be used with [docker-compose](#https://docs.docker.com/compose).

**NOTE**: Make sure to adjust the configuration according to your needs. 

```yaml
version: "3.9"
services:
  xmrig-mo:
    image: thelolagemann/xmrig-mo
    ports:
    - "3000:3000"
    - "3001:3001"
    volumes:
      - "$HOME/xmrig-mo:/cfg:rw"
    environment:
      - WALLET_ADDRESS: "88yUzYzB9wrR2r2o1TzXxDMENr6Kbadr3caqKTBUNFZ3dWVt6sJcpWBAwMwNRtEi7nHcBcqzmExNfdNK7ughaCeUFuXXpPp"
      - RIG_NAME: "docker-cpu"
```

## Notes

### User/Groups IDs
Often when using data volumes (`-v` flags) with docker, you will run into permissions issues that occur between the host
and container. See [here](https://medium.com/@mccode/understanding-how-uid-and-gid-work-in-docker-containers-c37a01d01cf)
for a more detailed breakdown of why this happens. 

To avoid these issues, specify the user ID and group ID that the application should run as by specifying the `PUID` and
`PGID` environment variables. By default, these are set to 1000:1000 which is generally the default UID/GID used
for the first non-system account created. If you are unsure as to the UID/GID of a username, run the following command:

`id <username>`

Which should output something like

`uid=1000(username) gid=1000(username) groups=1000(username),4(adm),27(sudo),119(lpadmin)
`

### Configuration
Although the environment variables should provide you with most of the configuration options needed to run this
container, there may be times when you want to alter the configuration of xmrig. This is as simple as locating the 
`config.json` file in your `/cfg` mount, making your wanted changes and saving the file. xmrig will automatically 
reload the configuration when it detects that the file has been modified.

### MSR
In order to apply MSR registers, the docker container must be run with the `--privileged` flag.

## Building
In order to build the container run the command.

```bash
docker build -f Dockerfile .
```

When building docker containers, you can pass build arguments with the `--build-arg` flag. Listed below are the available
build arguments you can pass during build.

| Argument | Description | Default |
| --- | --- | --- |
| `XMRIG_VERSION` | The version of xmrig-mo to build | `6.15.3-mo1` |
