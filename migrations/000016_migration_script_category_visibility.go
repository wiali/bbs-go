package migrations

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"

	"github.com/mlogclub/simple/sqls"
)

// migrate_category_visibility initializes historical categories as public.
func migrate_category_visibility() error {
	return sqls.WithTransaction(func(txCtx *sqls.TxContext) error {
		return txCtx.Tx.Model(&models.Category{}).
			Where("visibility IS NULL").
			Update("visibility", constants.CategoryVisibilityPublic).Error
	})
}
