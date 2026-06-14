

```bash
git clone https://github.com/5p2O5pav/v2wall.git

```

## 构建前端管理界面

```bash
cd /root/v2wall/web

# 安装依赖
npm install

# 生产构建
npm run build

```



## 移动整个 dist 目录
```bash
mkdir -p /root/v2wall/cmd/master/web

mv /root/v2wall/web/dist /root/v2wall/cmd/master/web/

```

## 编译二进制文件
```bash
cd /root/v2wall
go mod tidy
go mod download

```


```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/v2wall-master ./cmd/master

```

