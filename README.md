# drlog

`easylog` 是一个基于 [lumberjack](https://github.com/natefinch/lumberjack) 和 [zap](https://github.com/uber-go/zap) 封装的的golang日志库, 已在生产环境使用。


## How To Use

### 直接使用

```golang
import (
	"github.com/logerror/easylog"
	"github.com/logerror/easylog/pkg/option"
)
```

```golang
easylog.Warn("123")
```

InitGlobalLogger接收一个可变参数，你可以根据需求配置
### 定义日志级别
```golang
log := easylog.InitGlobalLogger(option.WithLogLevel("error"))
defer log.Sync()

easylog.Info("some error to log")
```

### 配置日志文件路径，大小
支持配置文件路径，文件大小，是否归档压缩日志文件
```golang
log := easylog.InitGlobalLogger(
	option.WithLogLevel("info"), 
	option.WithLogFile("2.log", 1, false),
)
defer log.Sync()

easylog.Info(" some error to log")
```
### 如果在将日志写入到文件是不希望在控制台同步输出，则使用option.WithConsole参数
```golang
logger = easylog.InitGlobalLogger(
    option.WithLogLevel("info"),
    option.WithLogFile(logFilePath, 5, false),
    option.WithConsole(false),
)
```

### 如果你需要在使用过程中改变日志的参数可以重新 InitGlobalLogger 然后将返回参数作为ReplaceLogger的参数
```golang
easylog.ReplaceLogger(logger)
```

