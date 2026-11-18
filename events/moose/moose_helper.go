//go:build moose

package moose

import (
	"fmt"
	"slices"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"

	moose "moose/events"
)

// autoConnectTargetChanged reports whether the auto-connect target differs between
// two configs.
func autoConnectTargetChanged(prev, curr config.AutoConnectData) bool {
	return prev.ServerTag != curr.ServerTag ||
		prev.Country != curr.Country ||
		prev.City != curr.City ||
		prev.Group != curr.Group
}

// getTokenRenewDate extracts the token renewal date for the current user from config.
// Returns empty string if config is nil or token data is not available.
func getTokenRenewDate(cfg *config.Config) string {
	if cfg == nil || cfg.TokensData == nil {
		return ""
	}

	tokenData, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return ""
	}

	return tokenData.TokenRenewDate
}

// uiItemType maps a UiItemsAction item type string to its moose enum.
func uiItemType(itemType string) moose.NordvpnappUserInterfaceItemType {
	switch itemType {
	case "textbox":
		return moose.NordvpnappUserInterfaceItemTypeTextBox
	case "show":
		return moose.NordvpnappUserInterfaceItemTypeTabOrScreen
	}
	return moose.NordvpnappUserInterfaceItemTypeButton
}

// sensitiveServerGroups lists server groups whose connections require stripping
// sensitive fields from analytics.
var sensitiveServerGroups = []config.ServerGroup{
	config.ServerGroup_DEDICATED_IP,
	config.ServerGroup_DEDICATED_SERVER,
}

// hasSensitiveServerGroup reports whether any group the chosen server belongs to
// is on the sensitive list.
// Note. Independent of the user-requested group.
func hasSensitiveServerGroup(groups []config.ServerGroup) bool {
	for _, g := range groups {
		if slices.Contains(sensitiveServerGroups, g) {
			return true
		}
	}
	return false
}

// hasDedicatedServerServices reports whether the user has an active dedicated server service.
func hasDedicatedServerService(services core.ServicesResponse) bool {
	expirationChecker := internal.SystemTimeExpirationChecker{}
	return slices.ContainsFunc(services,
		func(sd core.ServiceData) bool {
			return sd.Service.ID == auth.DedicatedServersServiceID && !expirationChecker.IsExpired(sd.ExpiresAt)
		},
	)
}

// TODO (LVPN-11418) - Replace with purchases
func findPayment(payments []core.PaymentResponse) (core.Payment, bool) {
	// Sort by CreatedAt descending
	slices.SortFunc(payments, func(a core.PaymentResponse, b core.PaymentResponse) int {
		return -a.Payment.CreatedAt.Compare(b.Payment.CreatedAt)
	})

	// Find first element matching criteria
	index := slices.IndexFunc(payments, isPaymentValid)
	if index < 0 {
		return core.Payment{}, false
	}

	return payments[index].Payment, true
}

// TODO (LVPN-11418) - Replace with purchases
func findOrder(p core.Payment, orders []core.Order) (core.Order, bool) {
	// Find order matching the payment
	if p.Subscription.MerchantID != 25 && p.Subscription.MerchantID != 3 {
		return core.Order{}, false
	}
	index := slices.IndexFunc(orders, func(o core.Order) bool {
		var cmpID int
		switch p.Subscription.MerchantID {
		case 3:
			cmpID = o.ID
		case 25:
			cmpID = o.RemoteID
		}
		return p.Payer.OrderID == cmpID
	})
	if index < 0 {
		return core.Order{}, false
	}

	return orders[index], true
}

// isPaymentValid reports whether the payment has a valid status
func isPaymentValid(pr core.PaymentResponse) bool {
	return isAcceptedPaymentStatus(pr.Payment.Status)
}

func isAcceptedPaymentStatus(status string) bool {
	return status == "done" ||
		status == "chargeback" ||
		status == "refunded" ||
		status == "partially_refunded" ||
		status == "trial"
}

// countFunc returns a number of elements in slice matching criteria
func countFunc[S ~[]E, E any](s S, f func(E) bool, stopAt int) int {
	count := 0
	for _, e := range s {
		if f(e) {
			count++
		}
		if stopAt >= 0 && count >= stopAt {
			return count
		}
	}
	return count
}

// vpnConnReasonToInternalType converts the VPN connection reason to the internal connection trigger
func vpnConnReasonToInternalType(t events.VPNConnectionReason) moose.NordvpnappVpnConnectionTrigger {
	switch t {
	case events.VPNConnectionReasonServerMaintenance:
		return moose.NordvpnappVpnConnectionTriggerServerMaintenance
	case events.VPNConnectionReasonAutoConnect:
		return moose.NordvpnappVpnConnectionTriggerAutoConnectUserSetting
	default:
		return moose.NordvpnappVpnConnectionTriggerNone
	}
}

// eventTriggerDomainToInternalType converts the domain-specific event trigger type to the internal
// representation
func eventTriggerDomainToInternalType(trigger events.TypeEventTrigger) moose.NordvpnappEventTrigger {
	switch trigger {
	case events.TriggerUser:
		return moose.NordvpnappEventTriggerUser
	case events.TriggerApp:
		fallthrough
	default:
		return moose.NordvpnappEventTriggerApp
	}
}

// eventStatusToInternalType converts the event status type to the internal representation
func eventStatusToInternalType(status events.TypeEventStatus) moose.NordvpnappEventStatus {
	switch status {
	case events.StatusSuccess:
		return moose.NordvpnappEventStatusSuccess
	case events.StatusFailure:
		return moose.NordvpnappEventStatusFailureDueToRuntimeException
	case events.StatusCanceled:
		return moose.NordvpnappEventStatusFailureDueToUserInterrupt
	case events.StatusAttempt:
		fallthrough
	default:
		return moose.NordvpnappEventStatusAttempt
	}
}

// connectionProtocolToInternalType converts the vpn protocol to the moose protocol, with NordLynx as UDP
func connectionProtocolToInternalType(p config.VPNProtocol) moose.NordvpnappVpnConnectionProtocol {
	//exhaustive:ignore
	switch p {
	case config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP:
		return moose.NordvpnappVpnConnectionProtocolUdp
	case config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP:
		return moose.NordvpnappVpnConnectionProtocolTcp
	case config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
		return moose.NordvpnappVpnConnectionProtocolWebtunnel
	}
	return moose.NordvpnappVpnConnectionProtocolNone
}

// connectionTechnologyToInternalType converts the vpn protocol to the moose technology
func connectionTechnologyToInternalType(p config.VPNProtocol) moose.NordvpnappVpnConnectionTechnology {
	//exhaustive:ignore
	switch p {
	case config.VPNProtocol_VPN_PROTOCOL_NORDLYNX:
		return moose.NordvpnappVpnConnectionTechnologyNordlynx
	case config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP, config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP:
		return moose.NordvpnappVpnConnectionTechnologyOpenvpn
	case config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
		return moose.NordvpnappVpnConnectionTechnologyNordwhisper
	}
	return moose.NordvpnappVpnConnectionTechnologyNone
}

// serverListOriginToInternalType converts server list origin to the internal representation
func serverListOriginToInternalType(sourceFromApi bool) moose.NordvpnappServerListSource {
	if sourceFromApi {
		return moose.NordvpnappServerListSourceRecommendedByApi
	}

	return moose.NordvpnappServerListSourceLocallyCachedServerList
}

// serverSelectionRuleToInternalType converts server selection rule to the internal representation
func serverSelectionRuleToInternalType(rule config.ServerSelectionRule) moose.NordvpnappServerSelectionRule {
	switch rule {
	case config.ServerSelectionRule_RECOMMENDED:
		return moose.NordvpnappServerSelectionRuleRecommended
	case config.ServerSelectionRule_CITY:
		return moose.NordvpnappServerSelectionRuleCity
	case config.ServerSelectionRule_COUNTRY:
		return moose.NordvpnappServerSelectionRuleCountry
	case config.ServerSelectionRule_SPECIFIC_SERVER:
		return moose.NordvpnappServerSelectionRuleSpecificServer
	case config.ServerSelectionRule_GROUP:
		return moose.NordvpnappServerSelectionRuleSpecialtyServer
	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		return moose.NordvpnappServerSelectionRuleSpecialtyServerWithCountry
	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		return moose.NordvpnappServerSelectionRuleSpecialtyServerWithSpecificServer
	default:
		return moose.NordvpnappServerSelectionRuleNone
	}
}

// serverGroupToInternalType converts the internal config server group to the Moose representation
func serverGroupToInternalType(group config.ServerGroup) moose.NordvpnappServerGroup {
	switch group {
	case config.ServerGroup_DOUBLE_VPN:
		return moose.NordvpnappServerGroupDoubleVpn
	case config.ServerGroup_ONION_OVER_VPN:
		return moose.NordvpnappServerGroupOnionOverVpn
	case config.ServerGroup_DEDICATED_IP:
		return moose.NordvpnappServerGroupDedicatedIp
	case config.ServerGroup_STANDARD_VPN_SERVERS:
		return moose.NordvpnappServerGroupStandard
	case config.ServerGroup_NW_OBFUSCATED:
		return moose.NordvpnappServerGroupObfuscated
	case config.ServerGroup_DEDICATED_SERVER:
		return moose.NordvpnappServerGroupDedicatedServer
	default:
		return moose.NordvpnappServerGroupNone
	}
}

// isAutoConnectForFastestServer reports whether the auto-connect configuration has no specific target
func isAutoConnectForFastestServer(ac config.AutoConnectData) bool {
	return ac.ServerTag == "" &&
		ac.Country == "" &&
		ac.City == "" &&
		ac.Group == config.ServerGroup_UNDEFINED
}

// autoConnectConnectionPreference maps the auto-connect target to connection preference
func autoConnectConnectionPreference(ac config.AutoConnectData) moose.NordvpnappConnectionPreference {
	if isAutoConnectForFastestServer(ac) {
		return moose.NordvpnappConnectionPreferenceFastest
	}
	return moose.NordvpnappConnectionPreferenceSpecific
}

// realTimeProtectionToInternalType converts real time protection to the internal representation
func realTimeProtectionToInternalType(enabled bool) moose.NordvpnappOptBool {
	if enabled {
		return moose.NordvpnappOptBoolTrue
	}

	return moose.NordvpnappOptBoolFalse
}

// deviceTypeToInternalType converts device type to the internal representation
func deviceTypeToInternalType(deviceType sysinfo.SystemDeviceType) moose.NordvpnappDeviceType {
	var dt moose.NordvpnappDeviceType

	switch deviceType {
	case sysinfo.SystemDeviceTypeDesktop:
		dt = moose.NordvpnappDeviceTypeDesktop
	case sysinfo.SystemDeviceTypeServer:
		dt = moose.NordvpnappDeviceTypeServer
	case sysinfo.SystemDeviceTypeContainer:
		dt = moose.NordvpnappDeviceTypeVirtualMachine
	default:
		dt = moose.NordvpnappDeviceTypeUndefined
	}

	return dt
}

// durationToConnectionFunnel formats the pause duration time
func durationToConnectionFunnel(duration time.Duration) string {
	return fmt.Sprintf("pause:%d", int(duration.Minutes()))
}
