package fakes

type FakeExtractor struct {
	Extracted    []string
	ExtractPath  string
	ExtractError error
}

func NewFakeExtractor() *FakeExtractor {
	return &FakeExtractor{}
}

func (e *FakeExtractor) Extract(tarFilePath string) (string, error) {
	if e.Extracted == nil {
		e.Extracted = []string{}
	}
	e.Extracted = append(e.Extracted, tarFilePath)

	if e.ExtractError != nil {
		return "", e.ExtractError
	}

	if e.ExtractPath != "" {
		return e.ExtractPath, nil
	}

	return "/some/extracted/path", nil
}
