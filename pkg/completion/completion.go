package completion

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/text"
)

type PkgSynchronizer interface {
	SyncPackages(...string) []db.IPackage
}

type HttpGetter interface {
	Get(string) (*http.Response, error)
}

// Show provides completion info for shells
func Show(dbExecutor PkgSynchronizer, httpGet HttpGetter, aurURL, completionPath string, interval int, force bool) error {
	err := Update(dbExecutor, httpGet, aurURL, completionPath, interval, force)
	if err != nil {
		return err
	}

	in, err := os.OpenFile(completionPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer in.Close()

	_, out, _ := text.AllPorts()
	_, err = io.Copy(out, in)
	return err
}

// Update updates completion cache to be used by Complete
func Update(dbExecutor PkgSynchronizer, httpGet HttpGetter, aurURL, completionPath string, interval int, force bool) error {
	info, err := os.Stat(completionPath)

	if os.IsNotExist(err) || (interval != -1 && time.Since(info.ModTime()).Hours() >= float64(interval*24)) || force {
		errd := os.MkdirAll(filepath.Dir(completionPath), 0o755)
		if errd != nil {
			return errd
		}
		out, errf := os.Create(completionPath)
		if errf != nil {
			return errf
		}

		if createAURList(aurURL, httpGet, out) != nil {
			defer os.Remove(completionPath)
		}

		erra := createRepoList(dbExecutor, out)

		out.Close()
		return erra
	}

	return nil
}

// CreateAURList creates a new completion file
func createAURList(aurURL string, httpGet HttpGetter, out io.Writer) error {
	u, err := url.Parse(aurURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "packages.gz")

	resp, err := httpGet.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)

	scanner.Scan()
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		_, err = fmt.Fprintf(out, "%s\tAUR\n", text)
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}

// CreatePackageList appends Repo packages to completion cache
func createRepoList(dbExecutor PkgSynchronizer, out io.Writer) error {
	for _, pkg := range dbExecutor.SyncPackages() {
		_, err := fmt.Fprintf(out, "%s\t%s\n", pkg.Name(), pkg.DB().Name())
		if err != nil {
			return err
		}
	}
	return nil
}
