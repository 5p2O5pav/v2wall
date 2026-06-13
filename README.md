

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
mv web/dist cmd/master/web/

```



cd /root/v2wall
go get github.com/lionsoul2014/ip2region/binding/golang@latest
go mod tidy
go mod download

CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/v2wall-master ./cmd/master


