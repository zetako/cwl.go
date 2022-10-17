package cwl

type Importer interface {
  Load(string) ([]byte, error)
}

