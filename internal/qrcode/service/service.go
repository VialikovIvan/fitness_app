package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	mbspDomain "github.com/VialikovIvan/fitness_app/internal/membership/domain"
	myjwt "github.com/VialikovIvan/fitness_app/internal/pkg/jwt"
	qrDomain "github.com/VialikovIvan/fitness_app/internal/qrcode/domain"
	"github.com/skip2/go-qrcode"
)

const relativeAccessPath = "/access"

type MembershipRepository interface {
	HasActiveMembershipByID(ctx context.Context, userID int64) (bool, error)
}

type TokenManager interface {
	GenerateAccess(ctx context.Context, userID int64) (string, error)
	ValidateAccess(ctx context.Context, tokenStr string) (int64, string, error)
}

type Cache interface {
	GetAccessURL(ctx context.Context, userID string) (string, error)
	SetAccessURL(ctx context.Context, userID string, url string) error
	DeleteAccessURL(ctx context.Context, userID string) error
	MarkJTIUsed(ctx context.Context, jti string) (bool, error)
	SetAccessCooldown(ctx context.Context, userID string) error
	IsInCooldown(ctx context.Context, userID string) (bool, error)
}

type Service struct {
	membershipRepo MembershipRepository
	tokens         TokenManager
	cache          Cache
	logger         *slog.Logger
	baseURL        string
	muMap          map[int64]*sync.Mutex
	muMapLock      sync.Mutex
}

func New(r MembershipRepository, t TokenManager, c Cache, l *slog.Logger, baseURL string) *Service {
	return &Service{
		membershipRepo: r,
		tokens:         t,
		cache:          c,
		logger:         l.With("layer", "service"),
		baseURL:        baseURL,
		muMap:          make(map[int64]*sync.Mutex),
	}
}

func (s *Service) Generate(ctx context.Context, userID int64) (*qrcode.QRCode, error) {
	mu := s.getMutexByID(userID)

	mu.Lock()
	defer mu.Unlock()

	inCooldown, err := s.cache.IsInCooldown(ctx, strconv.FormatInt(userID, 10))
	if err != nil {
		return nil, fmt.Errorf("is in cooldown: %w", err)
	}
	if inCooldown {
		return nil, qrDomain.ErrUserIsInCooldown
	}

	accessURL, err := s.cache.GetAccessURL(ctx, strconv.FormatInt(userID, 10))
	if err == nil {
		qrCode, err := qrcode.New(accessURL, qrcode.Medium)
		if err != nil {
			return nil, fmt.Errorf("generate qr code: %w", err)
		}
		return qrCode, nil
	}
	if !errors.Is(err, myredis.ErrNoSuchKey) {
		return nil, fmt.Errorf("get access url: %w", err)
	}

	hasActive, err := s.membershipRepo.HasActiveMembershipByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("has active membership: %w", err)
	}
	if !hasActive {
		return nil, mbspDomain.ErrUserDoesNotHaveActiveMembership
	}

	token, err := s.tokens.GenerateAccess(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	accessURL = s.baseURL + relativeAccessPath + fmt.Sprintf("?token=%s", token)

	if err := s.cache.SetAccessURL(ctx, strconv.FormatInt(userID, 10), accessURL); err != nil {
		s.logger.Error("failed to cache access url", "err", err)
	}

	qrCode, err := qrcode.New(accessURL, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("generate qr code: %w", err)
	}

	return qrCode, nil
}

func (s *Service) AllowAccess(ctx context.Context, token string) (bool, error) {
	userID, jti, err := s.tokens.ValidateAccess(ctx, token)
	if err != nil {
		if errors.Is(err, myjwt.ErrInvalidToken) {
			return false, nil
		}
		return false, fmt.Errorf("validate access token: %w", err)
	}

	firstUse, err := s.cache.MarkJTIUsed(ctx, jti)
	if err != nil {
		return false, fmt.Errorf("is jti used: %w", err)
	}
	if !firstUse {
		return false, nil
	}

	hasActive, err := s.membershipRepo.HasActiveMembershipByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("has active membership: %w", err)
	}
	if !hasActive {
		return false, nil
	}

	if err := s.cache.DeleteAccessURL(ctx, strconv.FormatInt(userID, 10)); err != nil {
		s.logger.Error("failed to delete access url cache", "err", err)
	}

	if err := s.cache.SetAccessCooldown(ctx, strconv.FormatInt(userID, 10)); err != nil {
		s.logger.Error("failed to set access cooldown", "err", err)
	}

	return true, nil
}

func (s *Service) getMutexByID(id int64) *sync.Mutex {
	s.muMapLock.Lock()
	defer s.muMapLock.Unlock()
	if s.muMap[id] == nil {
		s.muMap[id] = &sync.Mutex{}
	}
	return s.muMap[id]
}
