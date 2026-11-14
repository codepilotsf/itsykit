package editstore

import (
	"fmt"
	"unicode/utf8"

	"itsy/internal/editor"
	"itsy/internal/media"
	"itsy/internal/site"
	"itsy/internal/snips"
)

func (s *Store) applyBlog(p *site.Project, edit editor.BlogEdit) error {
	if _, ok := p.Apps["blog"]; !ok {
		return fmt.Errorf("blog is not enabled")
	}
	if edit.Action != "save" && edit.Action != "publish" && edit.Action != "unpublish" {
		return fmt.Errorf("invalid blog action")
	}
	at := -1
	var existing *snips.BlogPost
	for i, v := range p.Content.Blog {
		if v.ID == edit.Post.ID {
			at = i
			copy := v
			existing = &copy
			break
		}
	}
	editor.ApplyBlogAction(&edit.Post, edit.Action, existing)
	body, err := snips.Sanitize(edit.Post.Body)
	if err != nil {
		return &editor.MediaValidationError{Message: err.Error()}
	}
	edit.Post.Body = body
	if edit.Upload != nil {
		v := edit.Upload
		if !snips.ValidID(v.ID) || v.ID != edit.Post.Cover || len(v.Data) == 0 || !utf8.ValidString(v.Name+v.Alt) || utf8.RuneCountInString(v.Name) > 200 || utf8.RuneCountInString(v.Alt) > 1000 {
			return fmt.Errorf("invalid blog cover upload")
		}
		if _, ok := p.Content.Media[v.ID]; ok {
			return fmt.Errorf("photo ID is already used")
		}
		photo, err := s.saveCoverPhoto(v.Data, v.Name, v.Alt)
		if err != nil {
			return err
		}
		if p.Content.Media == nil {
			p.Content.Media = map[string]media.Photo{}
		}
		p.Content.Media[v.ID] = photo
	}
	if at < 0 {
		p.Content.Blog = append(p.Content.Blog, edit.Post)
	} else {
		p.Content.Blog[at] = edit.Post
	}
	if err := p.Content.ValidateBlog(); err != nil {
		return &editor.MediaValidationError{Message: err.Error()}
	}
	return nil
}
