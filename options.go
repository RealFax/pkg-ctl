package pkgCtl

type Option func(*Context)

func WithLogger(logger Logger) Option {
	return func(c *Context) {
		c.logger = logger
	}
}

func WithValues(values map[string]any) Option {
	if values == nil {
		return nil
	}
	return func(c *Context) {
		c.values = values
	}
}
