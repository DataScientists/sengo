package profileentryrepository

import (
	"sheng-go-backend/ent"
	ur "sheng-go-backend/pkg/usecase/repository"
)

type profileentryRepository struct {
	client *ent.Client
}

func NewprofileentryRepository(client *ent.Client) ur.ProfileEntry {
	return &profileentryRepository{client}
}
