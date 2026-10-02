package access

import "github.com/safing/portmaster/spn/access/account"

// communityProfile describes local capabilities, without an account or tokens.
func communityProfile() *UserRecord {
	user := &UserRecord{User: &account.User{
		Username: "local", State: account.UserStateApproved,
		CurrentPlan: &account.Plan{Name: "Community", FeatureIDs: []account.FeatureID{
			account.FeatureHistory, account.FeatureBWVis, account.FeatureVPNCompat,
		}},
		View: &account.View{Message: "Local features are free. Configure Tor or WireGuard in routing settings."},
	}}
	user.SetKey(userRecordKey)
	user.UpdateMeta()
	return user
}
