package services

import (
	"testing"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
)

func TestCategoryServiceCanViewCategory(t *testing.T) {
	tests := []struct {
		name       string
		visibility constants.CategoryVisibility
		user       *models.User
		want       bool
	}{
		{
			name:       "public category allows anonymous visitors",
			visibility: constants.CategoryVisibilityPublic,
			want:       true,
		},
		{
			name:       "login category rejects anonymous visitors",
			visibility: constants.CategoryVisibilityLogin,
			want:       false,
		},
		{
			name:       "login category allows signed in users",
			visibility: constants.CategoryVisibilityLogin,
			user:       &models.User{Model: models.Model{Id: 1}},
			want:       true,
		},
		{
			name:       "owner category rejects regular users",
			visibility: constants.CategoryVisibilityOwner,
			user:       &models.User{Model: models.Model{Id: 1}},
			want:       false,
		},
		{
			name:       "owner category allows owner",
			visibility: constants.CategoryVisibilityOwner,
			user:       &models.User{Model: models.Model{Id: 1}, Roles: constants.RoleOwner},
			want:       true,
		},
		{
			name:       "unknown visibility is denied",
			visibility: constants.CategoryVisibility(99),
			user:       &models.User{Model: models.Model{Id: 1}, Roles: constants.RoleOwner},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := &models.Category{Visibility: tt.visibility}
			if got := CategoryService.CanViewCategory(tt.user, category); got != tt.want {
				t.Fatalf("CanViewCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}
