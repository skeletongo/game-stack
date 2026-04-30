package combat

import "context"

type Service interface {
	GetState(uid int64) (*CombatState, error)
	CalcDamage(attackerID, targetID int64, skillID int32) (int32, bool, error)
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) GetState(uid int64) (*CombatState, error) {
	return s.store.GetCombatState(context.Background(), uid)
}

func (s *service) CalcDamage(attackerID, targetID int64, skillID int32) (int32, bool, error) {
	state, err := s.store.GetCombatState(context.Background(), attackerID)
	if err != nil {
		return 0, false, err
	}

	for _, sk := range state.Skills {
		if sk.ID == skillID {
			crit := false
			damage := sk.Damage
			return damage, crit, nil
		}
	}
	return 0, false, nil
}
