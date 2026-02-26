# 返回内容的格式
本项目的API通常返回JSON，格式预览如下：

```JSON

{
        "success":true,
        "msg":"API Execution OK"
}

```

```JSON

{
        "success":false,
        "msg":"API Execution met a problem",
        "error_type":2,
        "debug":"Can not bind your JSON Request body"
}

```


任何API都会返回`success`和`msg`字段，错误时还会返回`error_type`字段，如果打开`Debug.APIVerbose`会返回`debug`字段。

关于msg的内容，可以看每个API在logic包中的开头注释，具体信息位于`logic/errors.go`

另外在回应头里有字段`X-Request-Id`，这是我们为每个HTTP请求生成的唯一ID，可以用来在日志里查找相应的信息