package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"repin/internal/context/domain"
)

const (
	defaultLimit = 9
	maxLimit     = 100
)

type postLister interface {
	List(ctx context.Context, page, limit int) ([]domain.Post, int, error)
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	GetByURL(ctx context.Context, url string) (*domain.Post, error)
	Prev(ctx context.Context, id int64) (*domain.Post, error)
	Next(ctx context.Context, id int64) (*domain.Post, error)
}

type mediaLister interface {
	ListByPostIDs(ctx context.Context, ids []int64) (map[int64][]domain.PostMedia, error)
}

type PostService struct {
	posts postLister
	media mediaLister
}

func NewPostService(posts postLister, media mediaLister) *PostService {
	return &PostService{posts: posts, media: media}
}

type PostPage struct {
	Posts []domain.Post
	Page  int
	Limit int
	Total int
}

func (s *PostService) List(ctx context.Context, page, limit int) (PostPage, error) {
	if page < 1 {
		page = 1
	}

	switch {
	case limit < 1:
		limit = defaultLimit
	case limit > maxLimit:
		limit = maxLimit
	}

	posts, total, err := s.posts.List(ctx, page, limit)
	if err != nil {
		return PostPage{}, fmt.Errorf("list posts: %w", err)
	}

	if err := s.attachMedia(ctx, posts); err != nil {
		return PostPage{}, err
	}

	return PostPage{Posts: posts, Page: page, Limit: limit, Total: total}, nil
}

type PostDetails struct {
	Post *domain.Post
	Prev *domain.Post
	Next *domain.Post
}

func (s *PostService) resolve(ctx context.Context, slug string) (*domain.Post, error) {
	if id, ok := leadingID(slug); ok {
		return s.posts.GetByID(ctx, id)
	}

	return s.posts.GetByURL(ctx, slug)
}

func leadingID(slug string) (int64, bool) {
	digits, _, _ := strings.Cut(slug, "-")

	id, err := strconv.ParseInt(digits, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func (s *PostService) Get(ctx context.Context, slug string) (PostDetails, error) {
	post, err := s.resolve(ctx, slug)
	if err != nil {
		return PostDetails{}, fmt.Errorf("get post: %w", err)
	}

	prev, err := s.adjacent(ctx, s.posts.Prev, post.ID)
	if err != nil {
		return PostDetails{}, err
	}

	next, err := s.adjacent(ctx, s.posts.Next, post.ID)
	if err != nil {
		return PostDetails{}, err
	}

	posts := []domain.Post{*post}
	if prev != nil {
		posts = append(posts, *prev)
	}

	if next != nil {
		posts = append(posts, *next)
	}

	if err := s.attachMedia(ctx, posts); err != nil {
		return PostDetails{}, err
	}

	details := PostDetails{Post: &posts[0]}

	i := 1
	if prev != nil {
		details.Prev = &posts[i]
		i++
	}

	if next != nil {
		details.Next = &posts[i]
	}

	return details, nil
}

func (s *PostService) adjacent(ctx context.Context, fetch func(context.Context, int64) (*domain.Post, error), id int64) (*domain.Post, error) {
	post, err := fetch(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("adjacent post: %w", err)
	}

	return post, nil
}

func (s *PostService) attachMedia(ctx context.Context, posts []domain.Post) error {
	if len(posts) == 0 {
		return nil
	}

	ids := make([]int64, len(posts))
	for i := range posts {
		ids[i] = posts[i].ID
	}

	media, err := s.media.ListByPostIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("list post media: %w", err)
	}

	for i := range posts {
		posts[i].Media = media[posts[i].ID]
	}

	return nil
}
