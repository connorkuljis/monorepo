package blog

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/connorkuljis/monorepo/ssg/pkg/post"
)

const (
	PublicDir = "public"
	PostsDir  = "posts"
	SourceDir = "posts"
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
	err := os.MkdirAll(b.PublicPostsDir, os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			fmt.Printf("Directory %s already exists\n", b.PublicPostsDir)
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

			if !post.Matter.Draft {
				posts = append(posts, post)
			}
		}
	}

	b.Posts = posts

	return nil
}

func (b *Blog) BuildHomePage() error {
	filename := "index.html"

	t, err := template.New(filename).ParseFiles("templates/base.html", "templates/index.html")
	if err != nil {
		return err
	}

	b.IndexPage = t

	return nil
}

func (b *Blog) Save() error {
	filename := "index.html"
	f, err := os.Create(filepath.Join(b.PublicDir, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	err = b.IndexPage.ExecuteTemplate(f, "base", map[string]any{"Posts": b.Posts})
	if err != nil {
		return err
	}

	for _, post := range b.Posts {
		err := post.Render(b.PublicPostsDir)
		if err != nil {
			return err
		}
	}

	return nil
}
