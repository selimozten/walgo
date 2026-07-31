package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ganbitlabs/walgo/internal/sui"
	"github.com/ganbitlabs/walgo/internal/ui"
	"github.com/ganbitlabs/walgo/internal/walrus"
)

// expiryWarningWindow is how close to expiration a site has to be before walgo
// nags about renewing it. Mainnet epochs last two weeks, so a week of runway is
// the point where a user should act; on testnet, where epochs last a day, the
// warning is effectively permanent, which is accurate.
const expiryWarningWindow = 7 * 24 * time.Hour

// printExpiryStatus reports how much life a deployed site has left.
func printExpiryStatus(expired, total int, earliest time.Time) {
	icons := ui.GetIcons()

	if expired > 0 {
		fmt.Printf("%s Storage: %d of %d resources expired\n", icons.Cross, expired, total)
		fmt.Printf("   Expired storage cannot be extended. Re-upload with:  walgo update --epochs <n>\n")
		fmt.Printf("   The site object keeps its ID; only the content has to be stored again.\n")
		return
	}

	if earliest.IsZero() {
		return // nothing to report: no owned blob objects, so no dates
	}

	remaining := time.Until(earliest)
	days := int(remaining.Hours() / 24)

	if remaining < expiryWarningWindow {
		fmt.Printf("%s Storage expires %s (in %d days)\n", icons.Warning, earliest.Format(time.DateOnly), days)
		fmt.Printf("   Renew before then:  walgo update --epochs <n>\n")
		return
	}

	fmt.Printf("%s Storage expires %s (in %d days)\n", icons.Check, earliest.Format(time.DateOnly), days)
}

// updateExpiryCheck describes the site being updated, so the warning can price
// a re-upload when the existing storage has expired.
type updateExpiryCheck struct {
	ObjectID  string
	Network   string
	Epochs    int
	SiteSize  int64
	FileCount int
	Verbose   bool
}

// checkExpiryBeforeUpdate warns about expired storage before an update runs.
//
// site-builder handles expired blobs by re-storing them, but that is a silent
// full-price upload of content the user believes is already stored, so it is
// worth saying out loud before the wallet is charged. Failures here are not
// fatal: the update itself is the source of truth.
func checkExpiryBeforeUpdate(check updateExpiryCheck) {
	icons := ui.GetIcons()

	expiry, err := walrus.GetSiteExpiry(check.ObjectID)
	if err != nil {
		if check.Verbose {
			fmt.Fprintf(os.Stderr, "%s Could not read current site expiry: %v\n", icons.Warning, err)
		}
		return
	}

	switch {
	case expiry.HasExpired():
		fmt.Printf("\n%s %d of %d resources have expired storage\n", icons.Warning, len(expiry.Expired), expiry.Total)
		fmt.Printf("   Expired storage cannot be extended, so those files are uploaded again\n")
		fmt.Printf("   at full price from the local build. The site keeps its object ID.\n")
		if estimate := estimateRestoreCost(check); estimate != "" {
			fmt.Printf("   Estimated re-upload cost: %s\n", estimate)
		}
		fmt.Printf("   Afterwards, clean up the dead blob objects:  walrus burn-blobs --all-expired\n")
	case !expiry.Earliest.IsZero() && time.Until(expiry.Earliest) < expiryWarningWindow:
		fmt.Printf("\n%s Storage expires %s; this update extends it by %d epoch(s)\n",
			icons.Info, expiry.Earliest.Format(time.DateOnly), check.Epochs)
	}
}

// actualUpdateCost reads what the wallet was just charged, so deployment records
// hold the real figure instead of an estimate. Returns "" when it cannot be read.
func actualUpdateCost(network string) string {
	wallet, err := sui.GetActiveAddress()
	if err != nil || wallet == "" {
		return ""
	}

	gasInfo, err := walrus.GetLatestTransactionGas(wallet, network)
	if err != nil {
		return ""
	}

	return gasInfo.Format()
}

// estimateRestoreCost prices storing the whole site again, which is what an
// update does once the previous storage has expired. Returns "" when the size
// is unknown or the estimate fails.
func estimateRestoreCost(check updateExpiryCheck) string {
	if check.SiteSize <= 0 {
		return ""
	}

	breakdown, err := walrus.CalculateCost(walrus.CostOptions{
		SiteSize:  check.SiteSize,
		Epochs:    check.Epochs,
		FileCount: check.FileCount,
		Network:   check.Network,
	})
	if err != nil {
		return ""
	}

	return fmt.Sprintf("~%.4f WAL + ~%.4f SUI", breakdown.TotalWAL, breakdown.GasCostSUI)
}
