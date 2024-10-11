package internal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/connorkuljis/monorepo/ssg/internal/page"
	"github.com/connorkuljis/monorepo/ssg/internal/util"
)

type Site struct {
	PublicDir      string
	PublicPostsDir string
	EnableDrafts   bool
	// TODO: move these out

	BlogPages []*page.BlogPage
	HomePage  *page.HomePage
}

func NewSite(enableDrafts bool) (Site, error) {
	return Site{
		PublicDir:      util.PublicDir,
		PublicPostsDir: filepath.Join(util.PublicDir, util.PostsDir),
		EnableDrafts:   enableDrafts,
	}, nil
}

// CreateNewPublicDir force creates a new empty /public directory.
func (s *Site) CreateNewPublicDir() error {
	err := os.RemoveAll(s.PublicDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(s.PublicPostsDir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// BundleStaticContentToPublicDir copies all files and directories in the /static folder into the /public directory
func (s *Site) BundleStaticContentToPublicDir() error {
	err := os.CopyFS(s.PublicDir, os.DirFS(util.StaticDir))
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			fmt.Printf("Directory %s already exists\n", util.StaticDir)
		} else {
			return err
		}
	}

	return nil
}

func (s *Site) ParseMarkdownPosts() error {
	files, err := os.ReadDir(util.SourceDir)
	if err != nil {
		return err
	}

	var blogPages []*page.BlogPage
	for _, file := range files {
		// parse the markdown files in the /posts directory
		path := filepath.Join(util.SourceDir, file.Name())
		if filepath.Ext(path) == ".md" {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			r := strings.NewReader(string(bytes))
			post, err := page.ParsePage(r)
			if err != nil {
				return err
			}

			if !post.Draft || s.EnableDrafts {
				blogPages = append(blogPages, post)
			}
		}
	}

	// Sort the slice by date
	sort.Slice(blogPages, func(i, j int) bool {
		return blogPages[i].Created.After(blogPages[j].Created)
	})

	s.BlogPages = blogPages

	return nil
}

func (s *Site) BuildHomePage() error {
	h := &page.HomePage{
		Filename:      "index.html",
		FeaturedPosts: s.BlogPages,
	}
	s.HomePage = h
	return nil
}

// Genereate generates the static html
func (site *Site) Generate() (int, error) {
	var count int

	for _, blogPage := range site.BlogPages {
		filename := filepath.Join("public", "posts", blogPage.Filename)
		f, err := os.Create(filename)
		if err != nil {
			return count, err
		}
		defer f.Close()

		err = blogPage.Generate(f)
		if err != nil {
			return count, err
		}
		count++
	}

	filename := filepath.Join("public", site.HomePage.Filename)
	f, err := os.Create(filename)
	if err != nil {
		return count, err
	}
	defer f.Close()
	err = site.HomePage.Generate(f)
	if err != nil {
		return count, err
	}
	count++

	return count, nil
}
