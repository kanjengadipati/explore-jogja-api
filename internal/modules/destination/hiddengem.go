package destination

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"time"

	"pleco-api/internal/cache"
)

// hiddenGemCategories is the set of destination categories that are eligible
// for the hidden gem selection. Culinary and heritage are excluded — a well-known
// restaurant or famous museum is not a "hidden" place to visit.
var hiddenGemCategories = map[string]bool{
	"nature":    true,
	"beach":     true,
	"adventure": true,
	"family":    true,
	"weekend":   true,
	"camping":   true,
	"hidden-gem": true, // explicitly tagged by admin
}

// SelectHiddenGems returns up to cache.HiddenGemCount external-IDs for this
// week's curated Hidden Gem set. Rules:
//
//  1. "exclude" destinations are never eligible, regardless of quality.
//  2. "pin" destinations always win a slot, ordered by HiddenGemPinnedAt ASC
//     (earliest pin wins a slot first when more than HiddenGemCount are pinned).
//  3. Remaining slots are filled from natural candidates:
//     - category must be in hiddenGemCategories (culinary/heritage excluded)
//     - rating >= 4.5 (quality floor)
//     - review_count >= 50 (enough reviews for a reliable rating signal)
//     - review_count < 1000 (genuinely under the radar — not mainstream yet)
//     Ranked rating DESC → review_count ASC, then lightly shuffled within a
//     2× pool using a weekly ISO-week seed so the list feels fresh week to week.
func SelectHiddenGems(all []Destination) []string {
	var pinned, candidates []Destination

	for _, d := range all {
		// Only published destinations with approved content are eligible.
		// Treat empty status as published (legacy/seed data compatibility).
		if d.Status != "" && d.Status != "published" {
			continue
		}
		if d.ContentStatus != "" && d.ContentStatus != "published" {
			continue
		}

		switch d.HiddenGemOverride {
		case "exclude":
			continue
		case "pin":
			pinned = append(pinned, d)
		default:
			cat := strings.ToLower(strings.TrimSpace(d.Category))
			if hiddenGemCategories[cat] &&
				d.Rating >= 4.5 &&
				d.ReviewCount >= 50 &&
				d.ReviewCount < 1000 {
				candidates = append(candidates, d)
			}
		}
	}

	// Pinned: sort by pin timestamp ASC so earliest-pinned wins first.
	sort.SliceStable(pinned, func(i, j int) bool {
		pi, pj := pinned[i].HiddenGemPinnedAt, pinned[j].HiddenGemPinnedAt
		if pi == nil && pj == nil {
			return false
		}
		if pi == nil {
			return false // nil sorts last
		}
		if pj == nil {
			return true
		}
		return pi.Before(*pj)
	})

	// Natural candidates: sort quality-first, then apply a light weekly shuffle
	// on a wider pool so the list rotates without being fully random.
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Rating != candidates[j].Rating {
			return candidates[i].Rating > candidates[j].Rating
		}
		return candidates[i].ReviewCount < candidates[j].ReviewCount
	})

	// Pool is 2× the remaining slots needed (after pinned) so the shuffle has
	// meaningful diversity while keeping quality-bounded.
	remainingSlots := cache.HiddenGemCount - len(pinned)
	if remainingSlots < 0 {
		remainingSlots = 0
	}
	poolSize := remainingSlots * 2
	if poolSize > len(candidates) {
		poolSize = len(candidates)
	}
	pool := make([]Destination, poolSize)
	copy(pool, candidates[:poolSize])

	// Seed with ISO year+week — same value all week, rotates automatically.
	year, week := time.Now().ISOWeek()
	seed := int64(year*100 + week)
	r := rand.New(rand.NewSource(seed)) //nolint:gosec // non-cryptographic shuffle
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	// Assemble final list: pinned first, then shuffled natural candidates.
	ids := make([]string, 0, cache.HiddenGemCount)
	for _, d := range pinned {
		if len(ids) >= cache.HiddenGemCount {
			break
		}
		ids = append(ids, d.ExternalID)
	}
	for _, d := range pool {
		if len(ids) >= cache.HiddenGemCount {
			break
		}
		ids = append(ids, d.ExternalID)
	}
	return ids
}

// LoadHiddenGemIDs reads the weekly selection from cache, computing and
// caching it on first access after expiry (lazy-refresh, same pattern as
// loadTrendingIDs). Returns an empty map on any error — callers must handle
// the no-badge case gracefully.
func (s *Service) LoadHiddenGemIDs(ctx context.Context, cacheStore cache.Store) map[string]bool {
	var ids []string
	if ok, err := cacheStore.GetJSON(ctx, cache.KeyHiddenGemIDs(), &ids); err == nil && ok && len(ids) > 0 {
		return toIDSet(ids)
	}

	all, err := s.Repo.FindAllForCuration()
	if err != nil {
		return map[string]bool{}
	}

	ids = SelectHiddenGems(all)
	_ = cacheStore.SetJSON(ctx, cache.KeyHiddenGemIDs(), ids, cache.TTLHiddenGem)
	return toIDSet(ids)
}

func toIDSet(ids []string) map[string]bool {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}
