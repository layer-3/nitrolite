package action_gateway

import "github.com/layer-3/nitrolite/pkg/core"

// PermissiveActionAllower is an ActionAllower that allows all actions without any checks.
// It returns empty user allowances, indicating the user is not limited.
type PermissiveActionAllower struct{}

func NewPermissiveActionAllower() *PermissiveActionAllower {
	return &PermissiveActionAllower{}
}

func (p *PermissiveActionAllower) AllowAction(_ Store, _ string, _ core.GatedAction) error {
	return nil
}

func (p *PermissiveActionAllower) AllowAppRegistration(_ Store, _ string) error {
	return nil
}

func (p *PermissiveActionAllower) GetUserAllowances(_ Store, _ string) ([]core.ActionAllowance, error) {
	return []core.ActionAllowance{}, nil
}
