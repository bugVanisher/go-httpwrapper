## 什么是 go-httpwrapper？

如果你想快速实现http协议的分布式压测，那么go-httpwrapper将会是一个不错的选择！

[Boomer](https://github.com/myzhan/boomer) 是Locust框架worker端的go实现，它很好地弥补了Locust使用Python实现而导致性能不佳的缺陷。

go-httpwrapper对Boomer进行了http协议封装，只需要编写约定好格式的json字符串即可快速实现分布式的http压测。

go-httpwrapper非常适合用来实现压测平台的http协议压测，平台实现分布式逻辑和机器资源管理，用户只需要在Web页面写json即可。

## 安装

```shell
go get github.com/bugVanisher/go-httpwrapper
```

## 格式

约定好Json字符串格式：

```json
{
    "debug": true,
    "domain": "https://postman-echo.com",
    "header":{},
    "declare": ["{{ $sessionId := getSid }}"],
    "init_variables": {
        "roomId": 1001,
        "sessionId": "{{ $sessionId }}",
        "ids": "{{ $sessionId }}"
    },
    "running_variables": {
    	"tid": "{{ getRandomId 5000 }}"
    },
    "func_set": [
        {
            "key": "getTest",
            "method": "GET",
            "url": "/get?name=gannicus&roomId={{ .roomId }}&age=10&tid={{ .tid }}",
            "body": "{\"timeout\":10000}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .args.age (10 | toString )) }}"
        },
        {
            "key": "postTest",
            "method": "POST",
            "header":{
               "Cookie": "{{ .tid }}",
               "Content-Type": "application/json"
            },
            "url": "/post?name=gannicus",
            "body": "{\"timeout\":{{ .tid }}, \"retry\":true}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .data.timeout (.tid | toFloat64 ) ) (eq .data.retry false) }}"
        }
    ]
}
```

json字段说明：

- **debug**: 是否为debug模式，true/false，默认为false，如果为true，则在控制台打印请求日志
- **domain**: 目标地址，需要指定协议，如https。必须字段
- **header**: 设置header，一旦设置所有请求都会包含该header，格式：{"key1":"value1","key2":"value2"}，非必须字段
- **declare**: 声明变量，只初始化一次，可以被init_variables和running_variables引用，通常用在一些需要被多个变量同时引用的场景，格式：["{{ $sessionId := getSid }}", "{{ $timestamp := getTimeStamp }}"]。非必须字段
- **init_variables**: 初始化变量设置，在init_variables定义内部变量，可以通过函数的方式或者字符串，在接口的body、url、header做字符串替换，例如：定义"id": "{{ getRandomId }}"，在func_sets中的url中替换变量。另外，init_variables可以定义多个变量，并且只会初始化一次。非必须字段
- **running_variables**: 运行中变量设置，用法与init_variables较为类似，在接口的body、url和header做字符串替换，一般用于可变的**body**、**url或者header**。该接口在每次发请求前做替换，非必须字段
- **func_sets**: 压测接口的集合，数组类型，可以指定多个接口进行压测，**必须字段**
  - **key:** 指定压测接口的唯一标识，必须字段
  - **body**: 指定POST请求的body入参，可用**init_variables**或者**running_variables**中的变量替换，非必须字段
  - **method**: 请求的method方法，GET，PUT，POST等，需要大写，**必须字段**
  - **url**: 除域名以外的完整url，如果有GET参数，需要补充到url中，**必须字段**
  - **header**: 每个接口可以额外再增加需要的header，定义的方法与外层的**header**字段一致。如果字段与外层header相同，则覆盖外层字段值。非必须字段
  - **probability**: 执行接口函数的权重值，如果多个接口则按定义的比例去调用，所有概率和不需要满足100。**必须字段**
  - **validator:** 用于对请求响应报文进行校验。说明：这里的校验**区分类型**，其中返回结果json的数值型全部为float64，因此对比时需要注意。非必须字段

## 使用

将json字符串传入GetTaskList即可得到[]boomer.Task。

```
tasks := httpwrapper.GetTaskList(templateJsonStr)
boomer.Run(tasks...)
```

执行方式，同[boomer](https://github.com/myzhan/boomer#run)。

## 场景举例

#### 动态生成接口参数

压测平台可以在执行压测时通过变量传递以及模板函数定义的方式，实现动态生成接口参数的功能。

```json
{
  "declare": ["{{ $sessionId := getSid }}"],
  "init_variables": {
        "roomId": 1001,
        "sessionId": "{{ $sessionId }}",
    },
  "running_variables": {
    	"tid": "{{ getRandomId 5000 }}"
    },
	"func_sets": [
        {
            "key": "getTest",
            "method": "GET",
            "url": "/get?name=gannicus&roomId={{ .roomId }}&age=10&tid={{ .tid }}",
            "body": "{\"timeout\":10000}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .args.age (10 | toString )) }}"
        },
	]
}
```

declare、init_variables和running_variables内部模板函数说明（可在magic_func.go文件中定义）：

```
# 目前只实现了类型转换函数
toFloat64() # 转为Float64
toString() # 转为字符串
```

#### API混合压测

支持多个Http接口按比例混合压测，通过在func_sets数组内定义多个字典，指定probability字段，多个func_set之间的probability数值可以理解为各个func_set的压测比例，总和不需要为100。

```json
{
"func_set": [
        {
            "key": "getTest",
            "method": "GET",
            "url": "/get?name=gannicus&roomId={{ .roomId }}&age=10&tid={{ .tid }}",
            "body": "{\"timeout\":10000}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .args.age (10 | toString )) }}"
        },
        {
            "key": "postTest",
            "method": "POST",
            "header":{
               "Cookie": "{{ .tid }}",
               "Content-Type": "application/json"
            },
            "url": "/post?name=gannicus",
            "body": "{\"timeout\":{{ .tid }}, \"retry\":true}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .data.timeout (.tid | toFloat64 ) ) (eq .data.retry false) }}"
        }
    ]
}
```

#### 接口返回参数校验

###### 比较http状态码

```json
"validator": "{{ eq .http_status_code 200 }}"
```

###### 多字段比较

```json
"validator": "{{ and  (eq .http_status_code 200) (eq .enable false) }}"
```

###### 嵌套字段比较
```json
"validator": "{{ and  (eq .http_status_code 200) (eq .data.enable false) }}"
```
基于go模板库，访问变量变得非常简单，比如上面的.data.enable,它对应响应中的内容类似如下：

```json
{
	"data":{
		"enable": false
	}
}
```

这样我们就可以比较响应json中的任何字段了。

###### 使用模板函数

由于go是一种强类型语言，当将响应json字符串转为对象时，数值类型全部为float64，因此对比数值类型时需要进行类型转换，否则eq时会不通过，validator字段支持模板函数，因此可以将字段类型转换后比较。

```json
"validator": "{{ and  (eq .http_status_code 200) (eq .data.timeout (.tid | toFloat64 )) }}"
```



## LICENSE

Open source licensed under the MIT license (see *LICENSE* file for details).