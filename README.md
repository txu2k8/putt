# pzatest - A golang project for Stress | DevOps | Maintenance

If you have any questions or requirements, please let me know.
[tao.xu2008@outlook.com -- Tao.Xu](tao.xu2008@outlook.com)

## Install

```shell
go get -u gitlab.panzura.com/stress/pzatest
```

## Usage

### 1. Build local develop env

```shell
# 1.1 Get source code from git
git clone git@gitlab.panzura.com:stress/pzatest.git
cd pzatest

# 1.2 set env GOPATH
vi /etc/profile
export GOROOT=/usr/local/go  # Set the default goroot for go install path
export GOPATH=$HOME/workspace/gocode   # set default gopath for go src/pkgs path
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
source /etc/profile

# 1.3 Download third-part libs
go mod download
go mod tidy

# 1.4 build a binary for use(Not required)
go build

# 1.5 run with source code
go run main.go -h
./pzatest -h
```

### 2. Basic usage

```shell
$ ./pzatest.exe -h
2020-05-13T12:35:06 test INFO: Args: pzatest -h
pzatest project for "Stress | DevOps | Maintenance | ..."

Usage:
  pzatest [flags]
  pzatest [command]

Available Commands:
  deploy      Deploy test env
  stress      Vizion Stress test
  tools       Vizion ops tools
  help        Help about any command

Flags:
      --case stringArray         Test Case Array
      --dpl_group_ids ints       dpl group ids array (default [1])
  -h, --help                     help for pzatest
      --jd_group_ids ints        jd group ids array (default [1])
      --master_ips stringArray   Master nodes IP address Array
      --run_times int            Run test case with iteration loop (default 10)
      --ssh_key string           ssh login PrivateKey file full path
      --ssh_port int             ssh login port (default 22)
      --ssh_pwd string           ssh login password (default "password")
      --ssh_user string          ssh login user (default "root")
      --vset_ids ints            vset IDs array

Use "pzatest [command] --help" for more information about a command.
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
