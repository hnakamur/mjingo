package mjingo

// LoadFunc is the type of the function called when the engine is loading a template.
// A [Error] created with [NewErrorNotFound] should be returned when the template is not found.
type LoadFunc func(name string) (string, error)

type loaderStore struct {
	SyntaxConfig        syntaxConfig
	KeepTrailingNewline bool
	loader              LoadFunc
	templates           map[string]*compiledTemplate
}

func newLoaderStoreDefault() *loaderStore {
	return &loaderStore{
		SyntaxConfig: defaultSyntaxConfig,
		loader: func(name string) (string, error) {
			return "", NewErrorNotFound(name)
		},
		templates: make(map[string]*compiledTemplate),
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

func (s *loaderStore) get(name string) (*compiledTemplate, error) {
	t, ok := s.templates[name]
	if ok {
		return t, nil
	}
	source, err := s.loader(name)
	if err != nil {
		return nil, err
	}
	if err := s.insert(name, source); err != nil {
		return nil, err
	}
	return s.templates[name], nil
}

func (s *loaderStore) setLoader(f LoadFunc) {
	s.loader = f
}

type templateStore = loaderStore
