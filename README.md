# docker_pull

因为`docker pull`使用docker daemon拉取镜像, 且代理配置比较麻烦, 因此开发这个工具快速拉取镜像到本地。并可以使用`docker load`加载。

## CMD
```shell
docker_pull download # 下载镜像到本地
docker_pull pull     # 当本地有docker命令时 自动load镜像文件并删除 与download参数等同
docker_pull clean    # 清理缓存和配置
```

## Auth
1. 如果已经`docker login`登陆了 不需要指定token 会自动读取本地配置文件

2. 指定用户名和密码
```shell
docker_pull download -u $username -p $password ubuntu:20.04

docker_pull pull -u $username -p $password ubuntu:20.04
```

## Proxy
1. http proxy
```shell
docker_pull download --proxy socks5://127.0.0.1:1080 ubuntu:20.04

docker_pull pull --proxy socks5://127.0.0.1:1080 ubuntu:20.04
```

2. [lightsocks](https://github.com/gwuhaolin/lightsocks)
默认从`~/.lightsocks.json`读取配置
```shell
# 不指定配置
docker_pull download --lsocks ubuntu:20.04

# 指定配置
docker_pull download --lsocks --lsocks_path=/your/path/.lightsocks.json ubuntu:20.04


# 不指定配置
docker_pull pull --lsocks ubuntu:20.04

# 指定配置
docker_pull pull --lsocks --lsocks_path=/your/path/.lightsocks.json ubuntu:20.04
```

2. ssr
   默认从`~/.shadowsocks.json`读取配置
```shell
# 不指定配置
docker_pull download --ssr ubuntu:20.04

# 指定配置
docker_pull download --ssr --ssr_path=/your/path/.shadowsocks.json ubuntu:20.04

# 指定URL
docker_pull download --ssr --ssr_url=ssr://... ubuntu:20.04


# 不指定配置
docker_pull pull --ssr ubuntu:20.04

# 指定配置
docker_pull pull --ssr --ssr_path=/your/path/.shadowsocks.json ubuntu:20.04

# 指定URL
docker_pull pull --ssr --ssr_url=ssr://... ubuntu:20.04
```

## Load
```shell
docker_pull download ubuntu:20.04

docker load -i library_ubuntu_20.04.tar
```

```shell
# 当本地存在docker 命令时
docker_pull pull ubuntu:20.04
```

## Clean
```shell
# 清除cache
docker_pull clean

# 清除cache & 配置
docker_pull clean -a
```