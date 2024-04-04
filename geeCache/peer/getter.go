package peer

type Getter interface {
	Get(key string) ([]byte, error)
}

type NewGetter = func(string) Getter

type GetterFunc func(key string) ([]byte, error)

func (gf *GetterFunc) Get(key string) ([]byte, error) {
	return (*gf)(key)
}
