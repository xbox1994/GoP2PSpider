package types

import "fmt"

type Torrent struct {
	InfoHashHex string
	Name        string
	Length      int64
	Files       []*TFile
}

type TFile struct {
	Name   string
	Length int64
}

func (tf *TFile) String() string {
	return fmt.Sprintf("name: %s\n, size: %d\n", tf.Name, tf.Length)
}

func (t *Torrent) String() string {
	return fmt.Sprintf("link: %s\nname: %s\nsize: %d\nfile: %d\n",
		fmt.Sprintf("magnet:?xt=urn:btih:%s", t.InfoHashHex),
		t.Name,
		t.Length,
		len(t.Files),
	)
}
