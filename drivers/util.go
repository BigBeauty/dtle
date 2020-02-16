/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/hashicorp/nomad, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/actiontech/dtle/olddtle/internal/models"
)

type allocTuple struct {
	exist, updated *models.Allocation
}

// diffResult is used to return the sets that result from a diff
type diffResult struct {
	added   []*models.Allocation
	removed []*models.Allocation
	updated []allocTuple
	ignore  []*models.Allocation
}

func (d *diffResult) GoString() string {
	return fmt.Sprintf("allocs: (added %d) (removed %d) (updated %d) (ignore %d)",
		len(d.added), len(d.removed), len(d.updated), len(d.ignore))
}

// diffAllocs is used to diff the existing and updated allocations
// to see what has happened.
func diffAllocs(existing []*models.Allocation, allocs *allocUpdates) *diffResult {
	// Scan the existing allocations
	result := &diffResult{}
	existIdx := make(map[string]struct{})
	for _, exist := range existing {
		// Mark this as existing
		existIdx[exist.ID] = struct{}{}

		// Check if the alloc was updated or filtered because an update wasn't
		// needed.
		alloc, pulled := allocs.pulled[exist.ID]
		_, filtered := allocs.filtered[exist.ID]

		// If not updated or filtered, removed
		if !pulled && !filtered {
			result.removed = append(result.removed, exist)
			continue
		}

		// Check for an update
		if pulled && alloc.AllocModifyIndex > exist.AllocModifyIndex {
			result.updated = append(result.updated, allocTuple{exist, alloc})
			continue
		}

		// Ignore this
		result.ignore = append(result.ignore, exist)
	}

	// Scan the updated allocations for any that are new
	for id, pulled := range allocs.pulled {
		if pulled.DesiredStatus != models.AllocDesiredStatusPause {
			if _, ok := existIdx[id]; !ok {
				result.added = append(result.added, pulled)
			}
		}
	}
	return result
}

// persistState is used to help with saving state
func persistState(path string, data interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode state: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to make dirs for %s: %v", path, err)
	}
	tmpPath := path + ".tmp"
	if err := ioutil.WriteFile(tmpPath, buf, 0600); err != nil {
		return fmt.Errorf("failed to save state to tmp: %v", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename tmp to path: %v", err)
	}

	// Sanity check since users have reported empty state files on disk
	if stat, err := os.Stat(path); err != nil {
		return fmt.Errorf("unable to stat state file %s: %v", path, err)
	} else if stat.Size() == 0 {
		return fmt.Errorf("persisted invalid state file %s", path)
	}
	return nil
}
