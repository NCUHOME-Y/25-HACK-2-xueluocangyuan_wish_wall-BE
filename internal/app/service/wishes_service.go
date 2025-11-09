package service

import (
	"errors"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/repository"
)

type WishService interface {
	CreateWish(userID uint, content string, isPublic bool, tags []string, background string) (*model.Wish, error)
	GetPublicWishes(page, pageSize int, tag string) ([]model.Wish, int64, error)
	GetMyWishes(userID uint, page, pageSize int) ([]model.Wish, int64, error)
	UpdateWish(wishID, userID uint, updates map[string]interface{}) (*model.Wish, error)
	DeleteWish(wishID, userID uint) error
}

type wishService struct {
	wishRepo repository.WishRepository
	userRepo repository.UserRepository
}

func NewWishService(wishRepo repository.WishRepository, userRepo repository.UserRepository) WishService {
	return &wishService{
		wishRepo: wishRepo,
		userRepo: userRepo,
	}
}

func (s *wishService) CreateWish(userID uint, content string, isPublic bool, tags []string, background string) (*model.Wish, error) {
	// 参数验证
	if len(content) == 0 {
		return nil, errors.New("愿望内容不能为空")
	}
	if len(content) > 500 {
		return nil, errors.New("愿望内容长度不能超过500个字符")
	}
	if len(tags) > 5 {
		return nil, errors.New("标签数量不能超过5个")
	}

	wish := &model.Wish{
		UserID:     userID,
		Content:    content,
		IsPublic:   isPublic,
		Background: background,
	}

	// 创建愿望
	createdWish, err := s.wishRepo.CreateWish(wish, tags)
	if err != nil {
		return nil, err
	}

	return createdWish, nil
}

func (s *wishService) GetPublicWishes(page, pageSize int, tag string) ([]model.Wish, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.wishRepo.GetPublicWishes(page, pageSize, tag)
}

// 其他方法实现...
