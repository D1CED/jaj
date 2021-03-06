package upgrade

import (
	"bytes"
	"os/exec"
	"strconv"
	"testing"
	"time"

	alpm "github.com/Jguer/go-alpm/v2"
	rpc "github.com/mikkeloscar/aur"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"

	"github.com/Jguer/yay/v10/pkg/db/mock"
	"github.com/Jguer/yay/v10/pkg/exe"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/vcs"
)

func Test_upAUR(t *testing.T) {
	type args struct {
		remote     []alpm.IPackage
		aurdata    map[string]*rpc.Pkg
		timeUpdate bool
	}
	tests := []struct {
		name string
		args args
		want []Upgrade
	}{
		{
			name: "No Updates",
			args: args{
				remote: []alpm.IPackage{
					&mock.Package{PName: "hello", PVersion: "2.0.0"},
					&mock.Package{PName: "local_pkg", PVersion: "1.1.0"},
					&mock.Package{PName: "ignored", PVersion: "1.0.0", PShouldIgnore: true},
				},
				aurdata: map[string]*rpc.Pkg{
					"hello":   {Version: "2.0.0", Name: "hello"},
					"ignored": {Version: "2.0.0", Name: "ignored"},
				},
				timeUpdate: false,
			},
			want: []Upgrade{},
		},
		{
			name: "Simple Update",
			args: args{
				remote:     []alpm.IPackage{&mock.Package{PName: "hello", PVersion: "2.0.0"}},
				aurdata:    map[string]*rpc.Pkg{"hello": {Version: "2.1.0", Name: "hello"}},
				timeUpdate: false,
			},
			want: []Upgrade{{Name: "hello", Repository: "aur", LocalVersion: "2.0.0", RemoteVersion: "2.1.0"}},
		},
		{
			name: "Time Update",
			args: args{
				remote:     []alpm.IPackage{&mock.Package{PName: "hello", PVersion: "2.0.0", PBuildDate: time.Now()}},
				aurdata:    map[string]*rpc.Pkg{"hello": {Version: "2.0.0", Name: "hello", LastModified: int(time.Now().AddDate(0, 0, 2).Unix())}},
				timeUpdate: true,
			},
			want: []Upgrade{Upgrade{Name: "hello", Repository: "aur", LocalVersion: "2.0.0", RemoteVersion: "2.0.0"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := &bytes.Buffer{}
			text.CaptureOutput(buf, nil, func() {
				got := UpAUR(tt.args.remote, tt.args.aurdata, tt.args.timeUpdate)
				assert.EqualValues(t, tt.want, got)
			})

			cupaloy.SnapshotT(t, buf.Bytes())
		})
	}
}

type MockRunner struct {
	Returned []string
	Index    int
	t        *testing.T
}

func (r *MockRunner) Show(cmd *exec.Cmd) error {
	return nil
}

func (r *MockRunner) Capture(cmd *exec.Cmd, timeout int64) (stdout, stderr string, err error) {
	i, _ := strconv.Atoi(cmd.Args[len(cmd.Args)-1])
	if i >= len(r.Returned) {
		r.t.Log(r.Returned)
		r.t.Log(cmd.Args)
		r.t.Log(i)
	}
	stdout = r.Returned[i]
	assert.Contains(r.t, cmd.Args, "ls-remote")
	return stdout, stderr, err
}

func Test_upDevel(t *testing.T) {

	cmdRunner := &MockRunner{
		Returned: []string{
			"7f4c277ce7149665d1c79b76ca8fbb832a65a03b	HEAD",
			"7f4c277ce7149665d1c79b76ca8fbb832a65a03b	HEAD",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa	HEAD",
			"cccccccccccccccccccccccccccccccccccccccc	HEAD",
			"991c5b4146fd27f4aacf4e3111258a848934aaa1	HEAD",
		},
	}

	cmdBuilder := &exe.GitBuilder{
		GitBin: "git",
	}

	type args struct {
		remote  []alpm.IPackage
		aurdata map[string]*rpc.Pkg
		cached  vcs.InfoStore
	}
	tests := []struct {
		name     string
		args     args
		want     []Upgrade
		finalLen int
	}{
		{
			name: "No Updates",
			args: args{
				cached: vcs.InfoStore{
					Runner:     cmdRunner,
					GitBuilder: cmdBuilder,
				},
				remote: []alpm.IPackage{
					&mock.Package{PName: "hello", PVersion: "2.0.0"},
					&mock.Package{PName: "local_pkg", PVersion: "1.1.0"},
					&mock.Package{PName: "ignored", PVersion: "1.0.0", PShouldIgnore: true},
				},
				aurdata: map[string]*rpc.Pkg{
					"hello":   {Version: "2.0.0", Name: "hello"},
					"ignored": {Version: "2.0.0", Name: "ignored"},
				},
			},
			want: []Upgrade{},
		},
		{
			name:     "Simple Update",
			finalLen: 3,
			args: args{
				cached: vcs.InfoStore{
					Runner:     cmdRunner,
					GitBuilder: cmdBuilder,
					OriginsByPackage: map[string]vcs.OriginInfoByURL{
						"hello": {
							"github.com/Jguer/z.git": vcs.OriginInfo{
								Protocols: []string{"https"},
								Branch:    "0",
								SHA:       "991c5b4146fd27f4aacf4e3111258a848934aaa1",
							},
						},
						"hello-non-existant": {
							"github.com/Jguer/y.git": vcs.OriginInfo{
								Protocols: []string{"https"},
								Branch:    "0",
								SHA:       "991c5b4146fd27f4aacf4e3111258a848934aaa1",
							},
						},
						"hello2": {
							"github.com/Jguer/a.git": vcs.OriginInfo{
								Protocols: []string{"https"},
								Branch:    "1",
								SHA:       "7f4c277ce7149665d1c79b76ca8fbb832a65a03b",
							},
						},
						"hello4": {
							"github.com/Jguer/b.git": vcs.OriginInfo{
								Protocols: []string{"https"},
								Branch:    "2",
								SHA:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
							},
							"github.com/Jguer/c.git": vcs.OriginInfo{
								Protocols: []string{"https"},
								Branch:    "3",
								SHA:       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
							},
						},
					},
				},
				remote: []alpm.IPackage{
					&mock.Package{PName: "hello", PVersion: "2.0.0"},
					&mock.Package{PName: "hello2", PVersion: "3.0.0"},
					&mock.Package{PName: "hello4", PVersion: "4.0.0"},
				},
				aurdata: map[string]*rpc.Pkg{
					"hello":  {Version: "2.0.0", Name: "hello"},
					"hello2": {Version: "2.0.0", Name: "hello2"},
					"hello4": {Version: "2.0.0", Name: "hello4"},
				},
			},
			want: []Upgrade{
				Upgrade{
					Name:          "hello",
					Repository:    "devel",
					LocalVersion:  "2.0.0",
					RemoteVersion: "latest-commit",
				},
				Upgrade{
					Name:          "hello4",
					Repository:    "devel",
					LocalVersion:  "4.0.0",
					RemoteVersion: "latest-commit",
				},
			},
		},
		{
			name:     "No update returned",
			finalLen: 1,
			args: args{
				cached: vcs.InfoStore{
					Runner:     cmdRunner,
					GitBuilder: cmdBuilder,
					OriginsByPackage: map[string]vcs.OriginInfoByURL{
						"hello": {
							"github.com/Jguer/d.git": vcs.OriginInfo{
								Protocols: []string{"https"},
								Branch:    "4",
								SHA:       "991c5b4146fd27f4aacf4e3111258a848934aaa1",
							},
						},
					},
				},
				remote:  []alpm.IPackage{&mock.Package{PName: "hello", PVersion: "2.0.0"}},
				aurdata: map[string]*rpc.Pkg{"hello": {Version: "2.0.0", Name: "hello"}},
			},
			want: []Upgrade{},
		},
		{
			name:     "No update returned - ignored",
			finalLen: 1,
			args: args{
				cached: vcs.InfoStore{
					Runner:     cmdRunner,
					GitBuilder: cmdBuilder,
					OriginsByPackage: map[string]vcs.OriginInfoByURL{
						"hello": {
							"github.com/Jguer/e.git": vcs.OriginInfo{
								Protocols: []string{"https"},
								Branch:    "3",
								SHA:       "991c5b4146fd27f4aacf4e3111258a848934aaa1",
							},
						},
					},
				},
				remote:  []alpm.IPackage{&mock.Package{PName: "hello", PVersion: "2.0.0", PShouldIgnore: true}},
				aurdata: map[string]*rpc.Pkg{"hello": {Version: "2.0.0", Name: "hello"}},
			},
			want: []Upgrade{},
		},
	}

	text.CaptureOutput(nil, nil, func() {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.args.cached.Runner.(*MockRunner).t = t
				got := UpDevel(tt.args.remote, tt.args.aurdata, &tt.args.cached)
				assert.ElementsMatch(t, tt.want, got)
				assert.Equal(t, tt.finalLen, len(tt.args.cached.OriginsByPackage))
			})
		}
	})
}
