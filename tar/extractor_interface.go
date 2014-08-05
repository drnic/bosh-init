package tar

type Extractor interface {
	Extract(string) (string, error)
}
