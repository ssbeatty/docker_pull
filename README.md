# docker_pull

## Auth
1. 如果已经`docker login`登陆了 不需要指定token 会自动读取本地配置文件

2. 指定用户名和密码
```shell
docker_pull download -u $username -p $password ubuntu:20.04
```

## Proxy
1. http proxy
```shell
docker_pull download --proxy socks5://127.0.0.1:1080 ubuntu:20.04
```

2. [lightsocks](https://github.com/gwuhaolin/lightsocks)
默认从`~/.lightsocks.json`读取配置
```shell
# 不指定配置
docker_pull download --lsocks ubuntu:20.04

# 指定配置
docker_pull download --lsocks --lsocks_path=/your/path/.lightsocks.json ubuntu:20.04
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
```