# zio

CLI tool for AWS infrastructure

## Usage

```
Usage: zio [REGION] COMMAND [arg...]

Manage AWS infrastructure

Arguments:
  REGION="us-east-1"   AWS region ($AWS_REGION)

Options:
  -v, --version    Show the version and exit

Commands:
  instance, i   EC2 Instances
  reserved      Enumerate EC2 reserved instance status

Run 'zio COMMAND --help' for more information on a command.
```

### `instance`

```
Usage: zio instance [QUERY] [--stack=<stack name>] [--tag=<Name:Value>] COMMAND [arg...]

EC2 Instances

Arguments:
  QUERY=""     Fuzzy search query

Options:
  -s, --stack=""   Stack
  -t, --tag=""     Tag

Commands:
  exec, e      Execute command on instance
  ssh          SSH into an instance

Run 'zio instance COMMAND --help' for more information on a command.
```

#### Example

```
$ ./zio i curator-sandbox

+---------------------+--------------+--------------+----------+------------+-------------+-----------------------+
|     INSTANCE ID     |     NAME     |    STACK     |   TYPE   |     AZ     | IP ADDRESS  |       KEY NAME        |
+---------------------+--------------+--------------+----------+------------+-------------+-----------------------+
| i-0fb73a9d7ae6961f0 | nva1-curator | NVA1-Curator | m1.small | us-east-1d | 10.11.2.85  | zinc_sandbox_20161108 |
| i-0b965c297414d13d0 | nva1-curator | NVA1-Curator | m1.small | us-east-1a | 10.11.0.120 | zinc_sandbox_20161108 |
+---------------------+--------------+--------------+----------+------------+-------------+-----------------------+
```

### `instance ssh`

```
Usage: zio instance QUERY ssh

SSH into an instance
```

#### Example

```
$ ./zio i curator-sandbox ssh

Multiple instances found:

  1  nva1-curator  NVA1-Curator  10.11.2.85
  2  nva1-curator  NVA1-Curator  10.11.0.120

Login to [1]: 1
Welcome to Ubuntu 14.04.5 LTS (GNU/Linux 3.13.0-108-generic x86_64)
...
```

### `instance exec`

```
Usage: zio instance QUERY exec CMD [-c]

Execute command on instance

Arguments:
  CMD=""       Command to execute

Options:
  -c, --concurrency=2   Concurrency
```

#### Example

```
$ ./zio i curator-sandbox exec uptime -c 1

10.11.2.85      22:37:51 up 7 days, 23:48,  0 users,  load average: 0.25, 0.17, 0.20
10.11.0.120     22:37:51 up 7 days, 23:30,  0 users,  load average: 0.20, 0.16, 0.14
```

## Configuration

zio will look for `.ziorc` in the current directory or your home directory, in that order.

### Alias

Works much like the `git` alias option:

```
[alias]
  converge = exec "sudo chef-client" -c 3
```
