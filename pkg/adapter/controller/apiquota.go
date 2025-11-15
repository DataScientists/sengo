package controller

import (
	"context"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/usecase/usecase/apiquota"
)

type APIQuota interface {
	GetCurrent(ctx context.Context) (*ent.APIQuotaTracker, error)
	GetHistory(ctx context.Context, limit int) ([]*ent.APIQuotaTracker, error)
	SetOverride(ctx context.Context, enabled bool) (*ent.APIQuotaTracker, error)
	UpdateLimit(ctx context.Context, limit int) (*ent.APIQuotaTracker, error)
}

type apiQuotaController struct {
	quotaManager *apiquota.QuotaManager
}

func NewAPIQuotaController(qm *apiquota.QuotaManager) APIQuota {
	return &apiQuotaController{quotaManager: qm}
}

func (c *apiQuotaController) GetCurrent(ctx context.Context) (*ent.APIQuotaTracker, error) {
	return c.quotaManager.GetCurrentQuotaStatus(ctx)
}

func (c *apiQuotaController) GetHistory(ctx context.Context, limit int) ([]*ent.APIQuotaTracker, error) {
	return c.quotaManager.GetHistory(ctx, limit)
}

func (c *apiQuotaController) SetOverride(ctx context.Context, enabled bool) (*ent.APIQuotaTracker, error) {
	err := c.quotaManager.SetQuotaOverride(ctx, enabled)
	if err != nil {
		return nil, err
	}
	return c.quotaManager.GetCurrentQuotaStatus(ctx)
}

func (c *apiQuotaController) UpdateLimit(ctx context.Context, limit int) (*ent.APIQuotaTracker, error) {
	return c.quotaManager.UpdateLimit(ctx, limit)
}
