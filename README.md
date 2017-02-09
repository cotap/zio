# zio

CLI tool for zinc.io infrastructure

## Usage

```
Usage: zio [OPTIONS] COMMAND [arg...]

Manage zinc.io infrastructure

Options:
  -v, --version              Show the version and exit
  -r, --region="us-east-1"   AWS region ($AWS_REGION)

Commands:
  instance, instances, i   EC2 Instances

Run 'zio COMMAND --help' for more information on a command.
```

### `instance`

```
Usage: zio instance COMMAND [arg...]

EC2 Instances

Commands:
  list, ls     list instances
  ssh          Start SSH session

Run 'zio instance COMMAND --help' for more information on a command.
```

### `instance list`

```
Usage: zio instance list [OPTIONS]

list instances

Options:
  -e, --env=""    Filter by environment
  -r, --role=""   Filter by role
```

#### Example

```
$ ./zio i ls -e curator-sandbox

+---------------------+-----------------+---------+----------+------------+---------+-------------+-----------------------+
|     INSTANCE ID     |   ENVIRONMENT   |  ROLE   |   TYPE   |     AZ     |  STATE  | IP ADDRESS  |       KEY NAME        |
+---------------------+-----------------+---------+----------+------------+---------+-------------+-----------------------+
| i-0fb73a9d7ae6961f0 | curator-sandbox | curator | m1.small | us-east-1d | running | 10.11.2.85  | zinc_sandbox_20161108 |
| i-0b965c297414d13d0 | curator-sandbox | curator | m1.small | us-east-1a | running | 10.11.0.120 | zinc_sandbox_20161108 |
+---------------------+-----------------+---------+----------+------------+---------+-------------+-----------------------+
```

### `instance ssh`

```
Usage: zio instance ssh [OPTIONS]

Start SSH session

Options:
  -e, --env=""   Filter by environment
```
