package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/util/common"

	"gorm.io/gorm"
)

type inboundRuntime interface {
	AddInbound([]byte) error
	RemoveInbound(string) error
}

type inboundRuntimeChange struct {
	oldTag, newTag       string
	oldConfig, newConfig []byte
}

// Listeners cannot be prepared on ports still used by the old listener. Apply
// only this inbound, and restore it if activation or the database commit fails.
// ConfigService holds lifecycleMu throughout, including restoration.
func (change *inboundRuntimeChange) applyAndCommit(runtime inboundRuntime, commit func() error) (err error) {
	if runtime == nil {
		return commit()
	}
	oldRemoved, newAdded := false, false
	defer func() {
		if err == nil {
			return
		}
		if newAdded {
			if restoreErr := runtime.RemoveInbound(change.newTag); restoreErr != nil && !errors.Is(restoreErr, os.ErrInvalid) {
				err = errors.Join(err, fmt.Errorf("remove uncommitted inbound: %w", restoreErr))
			}
		}
		if oldRemoved && change.oldConfig != nil {
			if restoreErr := runtime.AddInbound(change.oldConfig); restoreErr != nil {
				err = errors.Join(err, fmt.Errorf("restore inbound %q: %w", change.oldTag, restoreErr))
			}
		}
	}()
	if change.oldTag != "" {
		// The manager removes the listener before closing its scope, so even a
		// scope-close error needs restoration. ErrInvalid means it was absent.
		oldRemoved = true
		if err = runtime.RemoveInbound(change.oldTag); errors.Is(err, os.ErrInvalid) {
			oldRemoved = false
		} else if err != nil {
			return err
		}
	}
	if change.newConfig != nil {
		if err = runtime.AddInbound(change.newConfig); err != nil {
			return err
		}
		newAdded = true
	}
	return commit()
}

func (s *InboundService) runtimeConfig(db *gorm.DB, inbound *model.Inbound) ([]byte, error) {
	if inbound == nil {
		return nil, nil
	}
	enabled, err := inbound.IsEnabled()
	if err != nil || !enabled {
		return nil, err
	}
	config, err := inbound.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return s.addUsers(db, config, inbound.Id, inbound.Type)
}

// Save database data, user bindings and the audit entry before touching the
// live listener. The switch uses a small patch to avoid overwriting a config
// that another panel request may have edited since the list was loaded.
func (s *InboundService) saveAndApply(db *gorm.DB, runtime inboundRuntime, act string, data json.RawMessage, initUsers, actor, hostname string) ([]string, error) {
	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()
	var ref struct {
		Id     uint   `json:"id"`
		Tag    string `json:"tag"`
		Enable *bool  `json:"enable"`
	}
	var before *model.Inbound
	switch act {
	case "new", "edit", "toggle":
		if err := json.Unmarshal(data, &ref); err != nil {
			return nil, err
		}
		if act == "new" {
			if ref.Id != 0 {
				return nil, common.NewError("new inbound id must be zero")
			}
		} else {
			if ref.Id == 0 || (act == "toggle" && ref.Enable == nil) {
				return nil, common.NewError("inbound id and boolean enable are required")
			}
			before = &model.Inbound{}
			if err := tx.Preload("Tls").First(before, ref.Id).Error; err != nil {
				return nil, err
			}
			if act == "toggle" {
				enabled, err := before.IsEnabled()
				if err != nil {
					return nil, err
				}
				if enabled == *ref.Enable {
					return []string{"inbounds"}, nil
				}
			}
		}
	case "del":
		if err := json.Unmarshal(data, &ref.Tag); err != nil {
			return nil, err
		}
		before = &model.Inbound{}
		if err := tx.Preload("Tls").Where("tag = ?", ref.Tag).First(before).Error; err != nil {
			return nil, err
		}
	default:
		return nil, common.NewErrorf("unknown action: %s", act)
	}
	change := &inboundRuntimeChange{}
	if before != nil {
		change.oldTag = before.Tag
	}
	if runtime != nil {
		var err error
		change.oldConfig, err = s.runtimeConfig(tx, before)
		if err != nil {
			return nil, err
		}
	}
	if err := s.Save(tx, act, data, initUsers, hostname); err != nil {
		return nil, err
	}
	if act != "del" && runtime != nil {
		var after model.Inbound
		query := tx.Preload("Tls")
		if act == "new" {
			query = query.Where("tag = ?", ref.Tag)
		} else {
			query = query.Where("id = ?", ref.Id)
		}
		if err := query.First(&after).Error; err != nil {
			return nil, err
		}
		change.newTag = after.Tag
		var err error
		change.newConfig, err = s.runtimeConfig(tx, &after)
		if err != nil {
			return nil, err
		}
		if act == "new" && change.newConfig != nil {
			change.newConfig, err = s.initUsers(tx, change.newConfig, initUsers, after.Type)
			if err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Create(&model.Changes{DateTime: time.Now().Unix(), Actor: actor, Key: "inbounds", Action: act, Obj: data}).Error; err != nil {
		return nil, err
	}
	if err := change.applyAndCommit(runtime, func() error { return tx.Commit().Error }); err != nil {
		return nil, err
	}
	if act == "toggle" {
		return []string{"inbounds"}, nil
	}
	return []string{"inbounds", "clients"}, nil
}
