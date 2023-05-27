## 错误处理

在处理业务逻辑时，如果出错误了，需要统一处理错误响应的格式，这样可以方便前端处理错误信息

所以需要定义一个 `Error` 接口，它包含了 `error` 接口，以及一个 `Status()` 方法，用来返回错误的状态码

```go
type Error interface {
	error
	Status() int
}
```

这个接口用来判断错误类型，在 `go` 中可以通过 `e.(type)` 判断错误的类型

```go
func respondWithError(w http.RespondWrite, err error) {
  switch e.(type) {
  case Error:
    respondWithJSON(w, e.Status(), e.Error())
  default:
    respondWithJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
  }
}
```

在 `go` 中 实现 `Error` 接口，只需要实现 `Error()` 和 `Status()` 方法即可

```go
func () Error() string {
  return ""
}
func () Status() int {
  return 0
}
```

这样定义的方法，只能返回固定的文本和状态码，如果想要返回动态内容，可以定义一个结构体

然后 `Error` 和 `Status` 方法接受 `StatusError` 类型

这样只要满足 `StatusError` 类型的结构体，就可以返回动态内容

所以上面的代码可以修改为：

```go
type StatusError struct {
  Code int
  Err error
}
func (se StatusError) Error() string {
  return se.Err.Error()
}
func (se StatusError) Status() int {
  return se.Code
}
```

## middlerware

### RecoverHandler

中间件 `RecoverHandler` 作用是通过 `defer` 来捕获 `panic`，然后返回 `500` 状态码

```go
func RecoverHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if r := recover(); r != nil {
        log.Println("Recover from panic %+v", r)
        http.Error(w, http.StatusText(500), 500)
      }
    }()
    next.ServeHTTP(w, r)
  }
  return http.HandlerFunc(fn)
}
```

### LoggingHandler

`LoggingHandler` 作用是记录请求耗时

```go
func (m Middleware) LoggingHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    next.ServeHTTP(w, r)
    end := time.Now()
    log.Printf("[%s] %q %v", r.Method, r.URL.Path, end.Sub(start))
  }
  return http.HandlerFunc(fn)
}
```

### 中间件使用

`alice` 是 `go` 中的一个中间件库，可以通过 `alice.New()` 来添加中间件，具体使用如下：

```go
m := alice.New(middleware.LoggingHandler, middleware.RecoverHandler)
mux.Router.HandleFunc("/api/v1/user", m.ThenFunc(controller)).Methods("POST")
```

## 生成短链接

### redis 连接

```go
func NewRedisCli(addr string, passwd string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})

	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}
	return &RedisCli{Cli: c}
}
```

### 生成唯一 ID

`redis` 可以基于一个键名生成一个唯一的自增 ID，这个键名可以是任意的，这个方法是 `Incr`

代码如下：

```go
err = r.Cli.Incr(URLIDKEY).Err()
if err != nil {
  return "", err
}

id, err := r.Cli.Get(URLIDKEY).Int64()
if err != nil {
  return "", err
}

fmt.Println(id) // 每次调用都会自增
```

### 存储和解析短链接

一个 `ID` 对应一个 `url`，也就是说当外面传入 `id` 时需要返回对应的 `url`

```go
func Shorten() {
  err := r.Cli.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
  if err != nil {
		return "", err
	}
}
func UnShorten() {
  url, err := r.Cli.Get(fmt.Sprintf(ShortlinkKey, eid)).Result()
}
```

### redis 注意事项

`redis` 返回的 `error` 有两种情况：

1. `redis.Nil` 表示没有找到对应的值
2. 其他错误，表示 `redis` 服务出错了

所以在使用 `redis` 时，需要判断返回的错误类型

```go
if err == redis.Nil {
  // 没有找到对应的值
} else if err != nil {
  // redis 服务出错了
} else {
  // 正确响应
}
```

## 测试

在测试用例中，如何发起一个请求，然后获取响应的数据呢？

1. 构造请求

```go
var jsonStr = []byte(`{"url":"https://www.baidu.com","expiration_in_minutes":60}`)
req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonStr))
if err != nil {
  t.Fatal(err)
}
req.Header.Set("Content-Type", "application/json")
```

2. 捕获 `http` 响应

```go
rw := httptest.NewRecorder()
```

3. 模拟请求被处理

```go
app.Router.ServeHTTP(rw, req)
```

4. 解析响应

```go
if rw.Code != http.ok {
  t.Fatalf("Excepted status created, got %d", rw.Code)
}

resp := struct {
  Shortlink string `json:"shortlink"`
}{}
if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
  t.Fatalf("should decode the response", err)
}
```

最终完整代码：

```go
var jsonStr = []byte(`{"url":"https://www.baidu.com","expiration_in_minutes":60}`)
req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonStr))
if err != nil {
  t.Fatal(err)
}
req.Header.Set("Content-Type", "application/json")

rw := httptest.NewRecorder()
app.Router.ServeHTTP(rw, req)

if rw.Code != http.ok {
  t.Fatalf("Excepted status created, got %d", rw.Code)
}
resp := struct {
  Shortlink string `json:"shortlink"`
}{}

if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
  t.Fatalf("should decode the response")
}
```

## 代码

### log.SetFlags(log.LstdFlags | log.Lshortfile)

作用是设置日志输出的标志

它们都是标志常量，用竖线 `|` 连接，这是位操作符，将他们合并为一个整数值，作为 `log.SetFlags()` 的参数

- `log.LstdFlags` 是标准时间格式：`2022-01-23 01:23:23`
- `log.Lshortfile` 是文件名和行号：`main.go:23`

当我们使用 `log.Println` 输出日志时，会自动带上时间、文件名、行号信息

### recover 函数使用

`recover` 函数类似于其他语言的 `try...catch`，用来捕获 `panic`，做一些处理

使用方法：

```go
func MyFunc() {
  defer func() {
    if r := recover(); r != nil {
      // 处理 panic 情况
    }
  }
}
```

需要注意的是：

1. `recover` 函数只能在 `defer` 中使用，如果在 `defer` 之外使用，会直接返回 `nil`
2. `recover` 函数只有在 `panic` 之后调用才会生效，如果在 `panic` 之前调用，也会直接返回 `nil`
3. `recover` 函数只能捕获当前 `goroutine` 的 `panic`，不能捕获其他 `goroutine` 的 `panic`

### next.ServerHttp(w, r)

`next.ServeHTTP(w, r)`，用于将 `http` 请求传递给下一个 `handler`

### HandleFunc 和 Handle 区别

`HandleFunc` 接受一个普通类型的函数：

```go
func myHandle(w http.ResponseWriter, r *http.Request) {}
http.HandleFunc("xxxx", myHandle)
```

`Handle` 接收一个实现 `Handler` 接口的函数：

```go
func myHandler(w http.ResponseWriter, r *http.Request) {}
http.Handle("xxxx", http.HandlerFunc(myHandler))
```

他们的区别是：使用 `Handle` 需要自己进行包装，使用 `HandleFunc` 不需要

### defer res.Body.Close()

为什么没有 `res.Header.Close()` 方法？

因为 `header` 不是资源，而 `body` 是资源，在 `go` 中，一般操作资源后，要及时关闭资源，所以 `go` 为 `body` 提供了 `Close()` 方法

`res.Body` 是 `io.ReadCloser` 类型的接口，表示可以读取响应数据并关闭响应体的对象

### w.Write()

代码在执行了 `w.Writer(res)` 后，还会继续往下执行，除非有显示的 `reture` 和 `panic` 终止函数执行

```go
func controller(w http.ResponseWriter, r *http.Request) {
  if res, err := xxx; err != nil {
    respondWithJSON(w, http.StatusOK, err)
  }
  // 这里如果有代码，会继续执行
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
  res, _ json.Marshal(payload)
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  w.Write(res)
}
```

需要注意的是，尽管执行了 `w.Writer()` 后，还会继续往下执行，但不会再对响应进行修改或写入任何内容了，因为 `w.Write()` 已经将响应写入到 `http.ResponseWriter` 中了

### 获取请求参数

#### 路由 /api/info?shortlink=2

```go
a.Router.Handle("/api/info", m.ThenFunc(a.getShortlinkInfo)).Methods("GET")

func getShortlinkInfo(w http.ResponseWriter, r *http.Request) {
  vals := r.URL.Query()
  s := vals.Get("shortlink")

  fmt.Println(s) // 2
}
```

#### 路由 /2

```go
a.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}}", m.ThenFunc(a.redirect)).Methods("GET")

func redirect(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  shortlink := vars["shortlink"]

  fmt.Println(shortlink) // 2
}
```

### 获取请求体

`json.NewDecoder(r.Body)` 作用是将 `http` 请求的 `body` 内容解析为 `json` 格式

`r.body` 是一个 `io.Reader` 类型，它代表请求的原始数据

如果关联成功可以用 `Decode()` 方法来解析 `json` 数据

```go
type User struct {
  Name string `json:"name"`
  Age int `json:"age"`
}

func controller(w http.ResponseWriter, r *http.Request){
  var user User
  if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
    fmt.Println(err)
  }
  fmt.Println(user)
}
```

### new

用于创建一个新的零值对象，并返回该对象的指针

它接受一个类型作为参数，并返回一个指向该类型的指针

适用于任何可分配的类型，如基本类型、结构体、数组、切片、映射和接口等

```go
// 创建一个新的 int 类型的零值对象，并返回指向它的指针
ptr := new(int)  // 0
```

需要注意的是：`new` 只分配了内存，并初始化为零值，并不会对对象进行任何进一步的初始化。如果需要对对象进行自定义的初始化操作，可以使用结构体字面量或构造函数等方式
