package mjingo

type loadFunc func(string) (string, error)

type loaderStore struct {
	SyntaxConfig        syntaxConfig
	KeepTrailingNewline bool
	loader              loadFunc
	templates           map[string]*compiledTemplate
}

func newLoaderStoreDefault() *loaderStore {
	return &loaderStore{
		SyntaxConfig: defaultSyntaxConfig,
		templates:    make(map[string]*compiledTemplate),
	}
}

func (s *loaderStore) insert(name, source string) error {
	t, err := newCompiledTemplate(name, source, s.SyntaxConfig, s.KeepTrailingNewline)
	if err != nil {
		return err
	}
	s.templates[name] = t
	return nil
}

func (s *loaderStore) remove(name string) {
	delete(s.templates, name)
}

func (s *loaderStore) clear() {
	clear(s.templates)
}

func (s *loaderStore) get(name string) *compiledTemplate {
	return s.templates[name]
}

func (s *loaderStore) setLoader(f loadFunc) {
	s.loader = f
}

type templateStore = loaderStore
