package site

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/connorkuljis/monorepo/ssg/internal/util"
)

type Site struct {
	EnableDrafts bool

	BlogPages []*BlogPage
	HomePage  *HomePage
}

func NewSite(enableDrafts bool) (Site, error) {
	return Site{EnableDrafts: enableDrafts}, nil
}

// CreateNewPublicDir force creates a new empty /public directory.
func (s *Site) CreateNewPublicDir() error {
	err := os.RemoveAll("public")
	if err != nil {
		return err
	}
	err = os.MkdirAll("public/posts", os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// BundleStaticContentToPublicDir copies all files and directories in the /static folder into the /public directory
func (s *Site) BundleStaticContentToPublicDir() error {
	err := os.CopyFS("public", os.DirFS("static"))
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			fmt.Printf("Directory %s already exists\n", util.StaticDir)
		} else {
			return err
		}
	}

	return nil
}

// 1. Read all files in posts directory.
// 2. For each file entry, check if it is a markdown file.
// 3. If it is a markdown file, open the file and read the contents into a sequence of bytes.
// 4. Parse the bytes to a blog post struct and exctract the post matter and content body.
// 5. Convert the markdown body to html.
// 6.
func (s *Site) ParseMarkdownPosts() error {
	files, err := os.ReadDir("posts")
	if err != nil {
		return err
	}

	var blogPages []*BlogPage
	for _, file := range files {
		// parse the markdown files in the /posts directory
		path := filepath.Join(util.SourceDir, file.Name())
		if filepath.Ext(path) == ".md" {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			r := strings.NewReader(string(bytes))
			post, err := ParsePage(r)
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
	h := &HomePage{
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
		filename := filepath.Join("public", "posts", blogPage.Slug)
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
