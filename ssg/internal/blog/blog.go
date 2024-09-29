package blog

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/connorkuljis/monorepo/ssg/internal/post"
)

const (
	PublicDir = "public"
	PostsDir  = "posts"
	SourceDir = "posts"
	StaticDir = "static"
)

type Blog struct {
	PublicDir      string
	PublicPostsDir string

	IndexPage *template.Template
	Posts     []post.Post
}

func NewBlog() (Blog, error) {
	var blog Blog

	blog.PublicDir = PublicDir
	blog.PublicPostsDir = filepath.Join(PublicDir, PostsDir)

	return blog, nil
}

func (b *Blog) Init() error {
	err := os.RemoveAll(b.PublicDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(b.PublicPostsDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.CopyFS(b.PublicDir, os.DirFS(StaticDir))
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			fmt.Printf("Directory %s already exists\n", StaticDir)
		} else {
			return err
		}
	}

	return nil
}

func (b *Blog) BuildPosts() error {
	files, err := os.ReadDir(SourceDir)
	if err != nil {
		return err
	}

	isMarkown := ".md"
	var posts []post.Post
	for _, f := range files {
		path := filepath.Join(SourceDir, f.Name())
		if filepath.Ext(path) == isMarkown {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			post, err := post.NewPost(bytes)
			if err != nil {
				return err
			}

			posts = append(posts, post)
		}
	}

	b.Posts = posts

	return nil
}

func (b *Blog) BuildHomePage() error {
	filename := "index.html"

	t, err := template.New(filename).ParseFiles(
		"templates/layout.html",
		"templates/head.html",
		"templates/view/index.html",
	)
	if err != nil {
		return err
	}

	b.IndexPage = t

	return nil
}

func (b *Blog) Save() (int, error) {
	var count int

	filename := "index.html"
	f, err := os.Create(filepath.Join(b.PublicDir, filename))
	if err != nil {
		return count, err
	}
	defer f.Close()

	err = b.IndexPage.ExecuteTemplate(f, "root", map[string]any{"Posts": b.Posts})
	if err != nil {
		return count, err
	}

	for _, post := range b.Posts {
		if !post.Matter.Draft {
			err := post.Render(b.PublicPostsDir)
			if err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}
