---
title: JSON Output
sidebar_position: 20
---

For `odo` to be used as a backend by graphical user interfaces (GUIs),
the useful commands can output their result in JSON format.

When used with the `-o json` flags, a command:
- that terminates successully, will:
  - terminate with a zero exit status,
  - will return its result in JSON format in its standard output stream.
- that terminates with an error, will:
  - terminate with a non-zero exit status,
  - will return an error message in its standard error stream, in the unique field `message` of a JSON object, as in `{ "message": "file not found" }`

The structures used to return information using JSON output are defined in [the `pkg/api` package](https://github.com/redhat-developer/odo/tree/main/pkg/api).

## odo analyze -o json

The `analyze` command analyzes the files in the current directory to select the best devfile to use,
from the devfiles in the registries defined in the list of preferred registries with the command `odo preference registry`.

The output of this command contains a devfile name and a registry name:

```bash
$ odo analyze -o json
{
    "devfile": "nodejs",
    "devfileRegistry": "DefaultDevfileRegistry"
}
$ echo $?
0
```

If the command is executed in an empty directory, it will return an error in the standard error stream and terminate with a non-zero exit status:

```bash
$ odo analyze -o json
{
	"message": "No valid devfile found for project in /home/user/my/empty/directory"
}
$ echo $?
1
```

## odo init -o json

The `init` command downloads a devfile and, optionally, a starter project. The usage for this command can be found in the [odo init command reference page](init.md).

The output of this command contains the path of the downloaded devfile and its content, in JSON format.

```bash
$ odo init -o json \
    --name aname \
    --devfile go \
    --starter go-starter
{
	"devfilePath": "/home/user/my-project/devfile.yaml",
	"devfileData": {
		"devfile": {
			"schemaVersion": "2.1.0",
      [...]
		},
		"supportedOdoFeatures": {
			"dev": true,
			"deploy": false,
			"debug": false
		}
	},
	"forwardedPorts": [],
	"runningIn": [],
	"managedBy": "odo"
}
$ echo $?
0
```

If the command fails, it will return an error in the standard error stream and terminate with a non-zero exit status:

```bash
# Executing the same command again will fail
$ odo init -o json \
    --name aname \
    --devfile go \
    --starter go-starter
{
	"message": "a devfile already exists in the current directory"
}
$ echo $?
1
```

## odo describe component -o json

The `describe component` command returns information about a component, either the component
defined by a Devfile in the current directory, or a deployed component given its name and namespace.

When the `describe component` command is executed without parameter from a directory containing a Devfile, it will return:
- information about the Devfile
  - the path of the Devfile,
  - the content of the Devfile,
  - supported `odo` features, indicating if the Devfile defines necessary information to run `odo dev`, `odo dev --debug` and `odo deploy`
- the status of the component
  - the forwarded ports if odo is currently running in Dev mode,
  - the modes in which the component is deployed (either none, Dev, Deploy or both)

```bash
$ odo describe component -o json
{
	"devfilePath": "/home/phmartin/Documents/tests/tmp/devfile.yaml",
	"devfileData": {
		"devfile": {
			"schemaVersion": "2.0.0",
			[ devfile.yaml file content ]
		},
		"supportedOdoFeatures": {
			"dev": true,
			"deploy": false,
			"debug": true
		}
	},
	"devForwardedPorts": [
		{
			"containerName": "runtime",
			"localAddress": "127.0.0.1",
			"localPort": 40001,
			"containerPort": 3000
		}
	],
	"runningIn": ["Dev"],
	"managedBy": "odo"
}
```

When the `describe component` commmand is executed with a name and namespace, it will return:
- the modes in which the component is deployed (either Dev, Deploy or both)

The command with name and namespace is not able to return information about a component that has not been deployed. 

The command with name and namespace will never return information about the Devfile, even if a Devfile is present in the current directory.

The command with name and namespace will never return information about the forwarded ports, as the information resides in the directory of the Devfile.

```bash
$ odo describe component --name aname -o json
{
	"runningIn": ["Dev"],
	"managedBy": "odo"
}
```

## odo list -o json

The `odo list` command returns information about components running on a specific namespace, and defined in the local Devfile, if any.

The `components` field lists the components either deployed in the cluster, or defined in the local Devfile.

The `componentInDevfile` field gives the name of the component present in the `components` list that is defined in the local Devfile, or is empty if no local Devfile is present.

In this example, the `component2` component is running in Deploy mode, and the command has been executed from a directory containing a Devfile defining a `component1` component, not running.

```bash
$ odo list --namespace project1
{
	"componentInDevfile": "component1",
	"components": [
		{
			"name": "component2",
			"managedBy": "odo",
			"runningIn": [
				"Deploy"
			],
			"projectType": "nodejs"
		},
		{
			"name": "component1",
			"managedBy": "",
			"runningIn": [],
			"projectType": "nodejs"
		}
	]
}
```

## odo registry -o json

The `odo registry` command lists all the Devfile stacks from Devfile registries. You can get the available flag in the [registry command reference](registry.md).

The default output will return information found into the registry index for stacks:

```shell
$ odo registry -o json
[
	{
		"name": "python-django",
		"displayName": "Django",
		"description": "Python3.7 with Django",
		"registry": {
			"name": "DefaultDevfileRegistry",
			"url": "https://registry.devfile.io",
			"secure": false
		},
		"language": "python",
		"tags": [
			"Python",
			"pip",
			"Django"
		],
		"projectType": "django",
		"version": "1.0.0",
		"starterProjects": [
			"django-example"
		]
	}, [...]
]
```

Using the `--details` flag, you will also get information about the Devfile:

```shell
$ odo registry --details -o json
[
	{
		"name": "python-django",
		"displayName": "Django",
		"description": "Python3.7 with Django",
		"registry": {
			"name": "DefaultDevfileRegistry",
			"url": "https://registry.devfile.io",
			"secure": false
		},
		"language": "python",
		"tags": [
			"Python",
			"pip",
			"Django"
		],
		"projectType": "django",
		"version": "1.0.0",
		"starterProjects": [
			"django-example"
		],
		"devfileData": {
			"devfile": {
				"schemaVersion": "2.0.0",
				[ devfile.yaml file content ]
			},
			"supportedOdoFeatures": {
				"dev": true,
				"deploy": false,
				"debug": true
			}
		},
	}, [...]
]

