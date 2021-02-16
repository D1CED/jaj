module github.com/D1CED/jaj

replace github.com/Jguer/yay/v10 => ./

require (
	github.com/Jguer/go-alpm/v2 v2.0.2
	github.com/Jguer/yay/v10 v10.0.0-00010101000000-000000000000
	github.com/Morganamilo/go-pacmanconf v0.0.0-20180910220353-9c5265e1b14f
	github.com/Morganamilo/go-srcinfo v1.0.0
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/leonelquinteros/gotext v1.4.0
	github.com/mikkeloscar/aur v0.0.0-20200113170522-1cb4e2949656
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20201207223542-d4d67f95c62d
	golang.org/x/term v0.0.0-20201207232118-ee85cb95a76b
	gopkg.in/h2non/gock.v1 v1.0.15
)

go 1.14
