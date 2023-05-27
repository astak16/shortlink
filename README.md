本项目是基于：[Go 开发短地址服务](https://www.imooc.com/learn/1150) 课程开发的

## 安装

```bash
git clone https://github.com/astak16/shortlink.git
cd shortlink
go mod tidy
```

## 运行

```bash
go run .
```

## API

### 生成短链接

> Api: /api/shorten
>
> Method: POST
>
> PARAMS: { "url": "https:www.example.com", "expire_in_minutes": 60 }

### 短链接跳转

> Api: /{shortlink}
>
> Method: GET

### 短链接详情

> Api: /api/info?info?shortlink={shortlink}
>
> Method: GET

## 项目笔记：

[go 开发地址服务笔记](./go开发地址服务笔记.md)
