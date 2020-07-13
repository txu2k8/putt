# A project for Maintenance | DevOps | StressTest (Golang)

If you have any questions or requirements, please let me know.
[tao.xu2008@outlook.com -- Tao.Xu](tao.xu2008@outlook.com)

## Install

```shell
go get -u gitlab.xxx.com/stress/platform
```

## Usage

### 1. Build local develop env

```shell
# 1.1 Get source code from git
git clone git@gitlab.xxx.com:stress/platform.git
cd platform

# 1.2 set env GOPATH
vi /etc/profile
export GOROOT=/usr/local/go  # Set the default goroot for go install path
export GOPATH=$HOME/workspace/go   # set default gopath for go src/pkgs path
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
source /etc/profile

# 1.3 Download third-part libs
go mod download
go mod tidy

# 1.4 build a binary for use(Not required)
go build

# 1.5 run with source code
go run main.go -h
./platform -h
platform.exe -h
```

### 2. Basic usage

```shell
$ ./platform.exe -h
2020-05-13T12:35:06 test INFO: Args: platform -h
platform project for "Maintenance | DevOps | StressTest ..."

Usage:
  platform [flags]
  platform [command]

Available Commands:
  deploy      Deploy test envmaint
  maint       Maintaince mode tools
  tools       DevOps tools
  stress      Stress test
  version     platform version
  help        Help about any command

Flags:
      --case stringArray         Test Case Array (default value in sub-command)
      --cmap_group_ids ints      cmap group ids array (default [1])
      --dpl_group_ids ints       dpl group ids array (default [1])
  -h, --help                     help for platform
      --jcache_group_ids ints    jcache group ids array (default [1])
      --jd_group_ids ints        jd group ids array (default [1])
      --k8s_namespace string     k8s namespace (default "vizion")
      --master_ips stringArray   Master nodes IP address Array
      --run_times int            Run test case with iteration loop (default 1)
      --ssh_key string           ssh login PrivateKey file full path (default "")
      --ssh_port int             ssh login port (default 22)
      --ssh_pwd string           ssh login password (default "password")
      --ssh_user string          ssh login user (default "root")
      --vset_ids ints            vset IDs array

Use "platform [command] --help" for more information about a command.
```

### 3. Module Details

#### 3.1 deploy

```shell
# Help to deploy env
```

#### 3.2 stress

```shell
# stress test cases
```

#### 3.3 tools

```shell
# DevOps tools
```
