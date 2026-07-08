package email

import (
	"context"
	"errors"
	"fmt"
)

var ErrAccountsExhausted = errors.New("all outreach email accounts exhausted for this run")

type AccountPool struct {
	providers         []Provider
	limitPerAccount   int
	currentIndex      int
	sentOnCurrent     int
	totalSent         int
	maxTotal          int
}

func NewAccountPool(providers []Provider, limitPerAccount, maxTotal int) (*AccountPool, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("account pool requires at least one email provider")
	}
	if limitPerAccount < 1 {
		return nil, fmt.Errorf("limit per account must be at least 1")
	}
	if maxTotal < 1 {
		return nil, fmt.Errorf("max total sends must be at least 1")
	}
	return &AccountPool{
		providers:       providers,
		limitPerAccount: limitPerAccount,
		maxTotal:        maxTotal,
	}, nil
}

func (pool *AccountPool) Exhausted() bool {
	return pool.currentIndex >= len(pool.providers) || pool.totalSent >= pool.maxTotal
}

func (pool *AccountPool) TotalSent() int {
	return pool.totalSent
}

func (pool *AccountPool) CurrentAccountIndex() int {
	if pool.currentIndex >= len(pool.providers) {
		return len(pool.providers)
	}
	return pool.currentIndex + 1
}

func (pool *AccountPool) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if pool.Exhausted() {
		return SendResult{}, ErrAccountsExhausted
	}

	for pool.currentIndex < len(pool.providers) {
		if pool.sentOnCurrent >= pool.limitPerAccount {
			pool.currentIndex++
			pool.sentOnCurrent = 0
			continue
		}
		if pool.totalSent >= pool.maxTotal {
			return SendResult{}, ErrAccountsExhausted
		}

		result, err := pool.providers[pool.currentIndex].Send(ctx, req)
		if err != nil {
			return SendResult{}, err
		}

		pool.sentOnCurrent++
		pool.totalSent++
		return result, nil
	}

	return SendResult{}, ErrAccountsExhausted
}

var _ Provider = (*AccountPool)(nil)
