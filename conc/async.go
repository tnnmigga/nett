package conc

import "github.com/tnnmigga/core/idef"

// 异步回调的方式执行函数
// 启动一个新的goruntine执行同步阻塞的代码
// 执行完将结果返到模块线程往后执行
// 匿名函数捕获的变量需要防范并发读写问题
func Async[T any](m idef.IModule, f func() (T, error), cb func(T, error)) {
	m.Async(func() (any, error) {
		res, err := f()
		return res, err
	}, func(a any, err error) {
		if a != nil {
			cb(a.(T), err)
		} else {
			var tmp T // 避免写ruturn nil导致的断言失败
			cb(tmp, err)
		}
	})
}
