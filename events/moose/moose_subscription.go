//go:build moose

package moose

import (
	"fmt"
	moose "moose/events"
	"strconv"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// formats the purchases according to NordvpnappSetContextUserNordvpnappSubscriptionHistory documentation
func formatPurchases(purchases []core.Purchase) string {
	var elements []string

	for _, p := range purchases {
		var plan core.PurchasePlan
		if len(p.Plans) == 0 {
			log.Warn("formatting purchases: no plan in purchases")
			continue
		}
		plan = p.Plans[0]

		subsById := make(map[int64]core.PurchaseSubscription, len(p.Subscriptions))
		for _, s := range p.Subscriptions {
			subsById[s.ID] = s
		}

		for _, pay := range p.Payments {
			if !isAcceptedPaymentStatus(pay.Status) {
				continue
			}

			sub, ok := subsById[pay.SubscriptionID]
			if !ok {
				log.Warn("formatting purchases: no subscription for payment")
				continue
			}

			t, err := time.Parse(internal.ServerDateFormat, pay.CreatedAt)
			if err != nil {
				log.Warn("formatting purchases: bad created_at for payment", err)
				continue
			}

			element := strings.Join([]string{
				t.Format(internal.YearMonthDateFormat),
				strconv.FormatInt(plan.ID, 10),
				strconv.FormatInt(sub.FrequencyInterval, 10),
				sub.FrequencyUnit,
				plan.Type,
				pay.Status,
				pay.Amount,
				pay.Currency,
				pay.Provider,
			}, ":")

			elements = append(elements, element)
		}
	}

	return strings.Join(elements, ",")
}

func (s *Subscriber) NotifyPurchases(purchases []core.Purchase) error {
	formattedData := formatPurchases(purchases)
	if err := s.response(moose.NordvpnappSetContextUserNordvpnappSubscriptionHistory(formattedData)); err != nil {
		return fmt.Errorf("setting subscription history %w", err)
	}
	return nil
}
