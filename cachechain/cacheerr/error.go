package cacheerr

import "errors"

var NoCacheSet = errors.New("没有配置任何缓存类型")
