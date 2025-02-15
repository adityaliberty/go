package resourceadapter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/guregu/null"
	"github.com/stellar/go/protocols/horizon/effects"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/support/render/hal"
	"github.com/stellar/go/support/test"
	"github.com/stretchr/testify/assert"
)

func TestNewEffectAllEffectsCovered(t *testing.T) {
	for typ, s := range EffectTypeNames {
		if typ == history.EffectAccountRemoved || typ == history.EffectAccountInflationDestinationUpdated {
			// these effects use the base representation
			continue
		}
		e := history.Effect{
			Type: typ,
		}
		result, err := NewEffect(context.TODO(), e, history.Ledger{})
		assert.NoError(t, err, s)
		// it shouldn't be a base type
		_, ok := result.(effects.Base)
		assert.False(t, ok, s)
	}

	// verify that the check works for an unknown effect
	e := history.Effect{
		Type: 20000,
	}
	result, err := NewEffect(context.TODO(), e, history.Ledger{})
	assert.NoError(t, err)
	_, ok := result.(effects.Base)
	assert.True(t, ok)
}

func TestEffectTypeNamesAreConsistentWithAdapterTypeNames(t *testing.T) {
	for typ, s := range EffectTypeNames {
		s2, ok := effects.EffectTypeNames[effects.EffectType(typ)]
		assert.True(t, ok, s)
		assert.Equal(t, s, s2)
	}
	for typ, s := range effects.EffectTypeNames {
		s2, ok := EffectTypeNames[history.EffectType(typ)]
		assert.True(t, ok, s)
		assert.Equal(t, s, s2)
	}
}

func TestNewEffect_EffectTrustlineAuthorizedToMaintainLiabilities(t *testing.T) {
	tt := assert.New(t)
	ctx, _ := test.ContextWithLogBuffer()

	details := `{
		"asset_code":   "COP",
		"asset_issuer": "GDRW375MAYR46ODGF2WGANQC2RRZL7O246DYHHCGWTV2RE7IHE2QUQLD",
		"asset_type":   "credit_alphanum4",
		"trustor":      "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3"
	}`

	hEffect := history.Effect{
		Account:            "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3",
		HistoryOperationID: 1,
		Order:              1,
		Type:               history.EffectTrustlineAuthorizedToMaintainLiabilities,
		DetailsString:      null.StringFrom(details),
	}
	resource, err := NewEffect(ctx, hEffect, history.Ledger{})
	tt.NoError(err)

	var resourcePage hal.Page
	resourcePage.Add(resource)

	effect, ok := resource.(effects.TrustlineAuthorizedToMaintainLiabilities)
	tt.True(ok)
	tt.Equal("trustline_authorized_to_maintain_liabilities", effect.Type)

	binary, err := json.Marshal(resourcePage)
	tt.NoError(err)

	var page effects.EffectsPage
	tt.NoError(json.Unmarshal(binary, &page))
	tt.Len(page.Embedded.Records, 1)
	tt.Equal(effect, page.Embedded.Records[0].(effects.TrustlineAuthorizedToMaintainLiabilities))
}
