package access

import (
	"github.com/safing/portmaster/spn/access/account"
	"testing"
)

func TestLocalFeaturesWithoutAccount(t *testing.T) {
	for _, user := range []*UserRecord{nil, {}, communityProfile()} {
		for _, feature := range []account.FeatureID{account.FeatureHistory, account.FeatureBWVis, account.FeatureVPNCompat} {
			if !user.MayUse(feature) {
				t.Fatalf("local feature %s requires an account", feature)
			}
		}
		if user.MayUseSPN() || user.MayUsePrioritySupport() {
			t.Fatal("local profile grants hosted services")
		}
	}
}
