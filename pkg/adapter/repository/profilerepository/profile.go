package profilerepository

import (
	"sheng-go-backend/ent"
	ur "sheng-go-backend/pkg/usecase/repository"
)

type profileRepository struct {
	client *ent.Client
}

func NewProfileRepository(client *ent.Client) ur.Profile {
	return &profileRepository{client}
}

