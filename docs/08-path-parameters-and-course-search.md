# 第八节 路径参数、课程详情与关键词筛选

上一篇：[第七节 net/http 与 GET 课程列表接口](./07-net-http-and-get-courses.md)。

本节在原来的课程列表接口上增加名称搜索，并通过一个新路由查询单门课程。所有数据仍然来自内存中的 courses 切片。

## 1 路径参数和查询参数

| 请求 | 参数类型 | 读取方式 |
| --- | --- | --- |
| GET /courses/1 | 路径参数 id | r.PathValue("id") |
| GET /courses?keyword=Go | 查询参数 keyword | r.URL.Query().Get("keyword") |

路径参数通常用于指定资源，例如 ID 为 1 的课程。查询参数通常用于描述查询条件，例如课程名称包含 Go。

`/courses?keyword=Go` 的路径仍然是 `/courses`，问号后面不是路径的一部分，因此不需要注册新路由，仍然进入已有的 `GET /courses` 处理函数。

`/courses/1` 则匹配新注册的 `GET /courses/{id}`。

## 2 注册课程详情路由

在 ListenAndServe 之前注册：

```go
mux.HandleFunc("GET /courses/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// 使用 id 查找课程并返回响应
})
```

`{id}` 是路径中的占位符。请求 `/courses/2` 时，id 的值是字符串 `"2"`，不是整数。它与 Course.ID 的 string 类型一致，不需要转换。

类似 React Router 的 `/courses/:id`，这里是在 Go 后端匹配和读取路径参数。

## 3 把查找函数连接到 HTTP 接口

之前学习的查找函数可以直接复用：

```go
func findCourseByID(courses []Course, id string) (Course, bool) {
	for _, course := range courses {
		if course.ID == id {
			return course, true
		}
	}
	return Course{}, false
}
```

处理函数中接住两个返回值：

```go
course, found := findCourseByID(courses, id)
if !found {
	http.Error(w, "课程不存在", http.StatusNotFound)
	return
}
```

course 是找到的课程值，found 表示是否找到。没有找到时，返回 HTTP 404，并用 return 结束处理，防止继续写入成功响应。

注意 http.Error 返回文本错误，不会自动包装成项目的 JSON 响应结构。

## 4 列表和详情的响应类型

当前代码分别定义了两个响应结构体：

```go
type ResData struct {
	Code    int      `json:"code"`
	Data    []Course `json:"data"`
	Message string   `json:"message"`
}

type FindCourseResData struct {
	Code    int    `json:"code"`
	Data    Course `json:"data"`
	Message string `json:"message"`
}
```

列表的 Data 是课程切片，JSON 中是数组。详情的 Data 是一门课程，JSON 中是对象。不能直接把一个 Course 赋值给 []Course 字段。

详情接口找到课程后：

```go
resData := FindCourseResData{
	Code:    200,
	Data:    course,
	Message: "请求成功！",
}

data, err := json.Marshal(resData)
if err != nil {
	http.Error(w, "课程数据编码失败", http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
_, err = w.Write(data)
if err != nil {
	log.Println("响应写入失败：", err)
}
```

请求 `/courses/2` 的成功响应：

```json
{
  "code": 200,
  "data": {"id": "2", "name": "学习Go切片", "completed": false},
  "message": "请求成功！"
}
```

## 5 每个接口都需要设置 Content-Type 吗

要根据实际响应内容决定，不能所有接口都写 application/json。

| 返回内容 | 常见 Content-Type |
| --- | --- |
| JSON | application/json |
| 普通文本 | text/plain; charset=utf-8 |
| HTML | text/html; charset=utf-8 |

当前两个接口的成功响应都是 JSON，所以分别设置 application/json。每次请求有自己的响应头，在列表接口中设置不会影响详情接口。

要在 WriteHeader 或 Write 之前设置响应头。不显式设置时，Go 会尝试判断内容类型，但不能依赖它将 JSON 自动标记为 application/json。

接口增多后可以考虑抽取公共响应函数；当前保持直接写法，先理解执行顺序。

## 6 从查询参数读取关键词

在已有列表处理函数开头写：

```go
keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
```

Query().Get 读取查询参数，缺少 keyword 时得到空字符串。TrimSpace 去掉两端空白，不会删除关键词中间的空格。

例如传入 `" 切片 "`，处理后是 `"切片"`；传入全空格，处理后是 `""`。

## 7 遍历并收集匹配课程

```go
filteredCourses := make([]Course, 0)

for _, course := range courses {
	if strings.Contains(course.Name, keyword) {
		filteredCourses = append(filteredCourses, course)
	}
}
```

make([]Course, 0) 创建长度为 0 的非 nil 切片。没有匹配结果时，它编码成 JSON 的 []；如果只声明 `var filteredCourses []Course` 并且没有追加元素，默认会编码成 null。

Contains 检查名称中是否包含关键词：

```go
strings.Contains("学习Go切片", "Go")   // true
strings.Contains("学习Go切片", "结构体") // false
strings.Contains("学习Go切片", "go")   // false
```

当前搜索区分大小写，不是拼音搜索，也不是自动忽略大小写的搜索。

任何字符串都包含空字符串，因此关键词为空时，全部课程都会被收集，不需要额外写一个分支。

## 8 append 这句做了什么

```go
filteredCourses = append(filteredCourses, course)
```

append 将当前 course 添加到结果切片末尾，返回添加后的切片描述，再赋值给 filteredCourses，更新它的长度等信息。

这和之前的 `courses = append(courses, newCourse)` 是相同语法。append 不保证每次创建新底层数组：容量足够时复用，不足时分配，因此要接住返回值。

当前 course 是 range 得到的结构体副本，追加到独立结果切片中，不会从原 courses 删除课程，也不会改写原课程数据。筛选结果顺序与原列表一致。

```text
结果初始为 []
遇到匹配课程 → append 加入结果
遇到不匹配课程 → 跳过
循环完成 → 返回结果切片
```

和前端这段代码的筛选目的相同，但 Go 的实现方式是循环收集：

```javascript
const filteredCourses = courses.filter(course =>
  course.name.includes(keyword),
)
```

## 9 必须把筛选结果用于响应

```go
resData := ResData{
	Code:    200,
	Data:    filteredCourses,
	Message: "请求成功！",
}
```

如果这里仍写 Data: courses，前面即使筛选正确，客户端收到的也仍然是全部课程。

没有搜索结果属于正常查询结果，返回 HTTP 200 和 data: []。这与详情接口指定一个不存在的 ID 返回 404 不同。

当前还保留了 findCourseByName 练习函数，但接口没有调用它。该函数用 == 精确匹配名称并返回第一门课程，不等于本节收集所有匹配项的包含搜索。

## 10 练习与本次验证

修改代码后停止旧服务，重新 go run .。依次访问：

| 请求 | 预期 |
| --- | --- |
| /courses | 全部三门课程 |
| /courses?keyword= | 全部三门课程 |
| /courses?keyword=Go | 全部三门课程 |
| /courses?keyword=切片 | 只有 ID 为 2 的课程 |
| /courses?keyword=不存在 | data 为 [] |
| /courses?keyword=go | data 为 []，大小写不同 |
| /courses/2 | data 为 ID 为 2 的课程对象 |
| /courses/99 | HTTP 404，课程不存在 |

终端中可以让 curl 编码中文和空格参数：

```sh
curl -i -G --data-urlencode 'keyword= 切片 ' http://localhost:8080/courses
```

本次检查使用当前源码的临时副本，仅替换监听地址为临时本机端口，实际请求验证了上述列表与详情行为，也检查了全空格及关键词两端空格的情况，没有打断已有服务。

编译和 go vet 通过。当前没有自动化测试文件，go test 显示 [no test files]；这些结果来自接口请求检查，不代表已有持久化的单元测试。未进行 React 前端联调。
