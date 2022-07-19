# gocron

> Cron是系统中用来定期执行或指定程序任务的一种服务或软件
>
> 它可以利用其实现周期性在特定日期/时间运行任务，是自动化运行重型任务的好工具，否则需要人工干预。

# 💡  简介

## Cron表达式

由空格分隔的六个字段组成

```xml
Seconds Minutes Hours Day Month Week
```

每个字段的含义如下:

| 字段名       | 取值范围        | 允许的特殊字符        |
| ------------ | --------------- | --------------------- |
| 秒 (seconds) | 0-59            | *    /    ,    -      |
| 分 (minutes) | 0-59            | *    /    ,    -      |
| 小时 (hours) | 0-23            | *    /    ,    -      |
| 日期 (day)   | 1-31            | *    /    ,    -    ? |
| 月份 (month) | 1-12 or JAN-DEC | *    /    ,    -      |
| 星期 (week)  | 0-6 or SUN-SAT  | *    /    ,    -    ? |

### 特殊字符

|      | 函数             | 说明                                                         |
| ---- | ---------------- | ------------------------------------------------------------ |
| *    | 表示匹配任意值   | 例如：在第五个字段(Month)中使用星号表示每个月                |
| /    | 表示范围的增量   | 例如：在第二个字段中使用`4-54/25`表示每小时的第4分钟开始到第54分钟，每隔25分钟执行 |
| ,    | 表示列出特定值   | 例如：在第五个字段中使用`MON,WED,SUN`表示在每周一，周三，周天执行 |
| -    | 表示范围         | 例如：在第三个字段中使用`12-15`表示每天中午12点到下午3点     |
| ?    | 表示忽略该字段值 | 例如：`15 30 12 9 * ?`表示每月9号的中午12:30:15，而不是每周的每天 |

# 🚀 功能

- 支持cron引擎的启动与停止
- 支持cron表达式解析

# 🌟 亮点

- 使用正则表达式匹配并解析cron表达式
- 使用优先队列进行调度

# 🎬 用法示例

```go
cron := New()
cron.Start()

err := cron.AddFunc("* * * * * ?", func() {
       fmt.Println("test")
})
if err != nil {
	fmt.Println(err)
}
cron.Stop()
```

# 📌 TODO

- [ ] 增加定时任务的选项(初始化函数, defer函数)
- [ ] 增加删除定时任务的功能
- [ ] 解决时间延迟问题
- [ ] 增加更方便的表达式
- [x] 增强性能: job调度用优先队列来实现

# 📔 参考文献

[Wiki](https://en.wikipedia.org/wiki/Cron) [CSDN](https://blog.csdn.net/darjun/article/details/106982893?ops_request_misc=%257B%2522request%255Fid%2522%253A%2522165820597916782350889519%2522%252C%2522scm%2522%253A%252220140713.130102334.pc%255Fall.%2522%257D&request_id=165820597916782350889519&biz_id=0&utm_medium=distribute.pc_search_result.none-task-blog-2~all~first_rank_ecpm_v1~pc_rank_34-15-106982893-null-null.142^v32^pc_rank_34,185^v2^control&utm_term=go%20cron&spm=1018.2226.3001.4187)

[cron](https://github.com/robfig/cron)

[cronexpr](https://github.com/gorhill/cronexpr)

[gron](https://github.com/roylee0704/gron)

[gods](https://github.com/emirpasic/gods)
