package health

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthService struct {
	pool *pgxpool.Pool
}

func NewHealthService(pool *pgxpool.Pool) *HealthService {
	return &HealthService{
		pool: pool,
	}
}

func (hs *HealthService) Ready(ctx context.Context) ReadyResponse {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := hs.pool.Ping(pingCtx); err != nil {
		hs.pool.Close()
		return ReadyResponse{
			Status: "unavailable",
			Dependecies: []DependecyStatus{
				{
					Name:   "postgres",
					Status: "down",
				},
			},
		}
	}

	return ReadyResponse{
		Status: "ready",
		Dependecies: []DependecyStatus{
			{
				Name:   "postgres",
				Status: "up",
			},
		},
	}
}
