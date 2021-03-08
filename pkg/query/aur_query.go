package query

import "github.com/mikkeloscar/aur"

type Pkg = aur.Pkg

type AUR struct {
	URL string
}

func (a AUR) Info(pkgs []string) ([]Pkg, error) {
	backup := aur.AURURL
	aur.AURURL = a.URL
	p, err := aur.Info(pkgs)
	aur.AURURL = backup
	return p, err
}
func (a AUR) Orphans() ([]Pkg, error) {
	backup := aur.AURURL
	aur.AURURL = a.URL
	p, err := aur.Orphans()
	aur.AURURL = backup
	return p, err
}
func (a AUR) Search(query string) ([]Pkg, error) {
	backup := aur.AURURL
	aur.AURURL = a.URL
	p, err := aur.Search(query)
	aur.AURURL = backup
	return p, err
}
func (a AUR) SearchBy(query string, by aur.By) ([]Pkg, error) {
	backup := aur.AURURL
	aur.AURURL = a.URL
	p, err := aur.SearchBy(query, by)
	aur.AURURL = backup
	return p, err
}
