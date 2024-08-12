package geeCache2

type LocalGetter interface {
	Get(key string) ([]byte, error)
}

type LocalGetterFn func(key string) ([]byte, error)

func (lg LocalGetterFn) Get(key string) ([]byte, error) {
	return lg(key)
}
